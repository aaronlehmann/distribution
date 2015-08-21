package storage

import (
	"errors"
	"io"
	"path"
	"sort"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

var _ distribution.TagService = &tagStore{}

// tagStore provides methods to manage manifest tags in a backend storage driver.
// This implementation uses the same on-disk layout as the (now deleted) signature
// store.  This provides backward compatibility with current registry deployments
// which only makes use of the Digest field of the returned distribution.Descriptor
// but does not enable full roundtripping of Descriptor objects
type tagStore struct {
	repository *repository
	blobStore  *blobStore
}

// Enumerate enumerates the tags stored in this tag store.
func (ts *tagStore) Enumerate(ctx context.Context, tags []string, last string) (n int, err error) {
	if len(tags) == 0 {
		return 0, errors.New("no space in slice")
	}

	pathSpec, err := pathFor(manifestTagPathSpec{
		name: ts.repository.Name(),
	})
	if err != nil {
		return 0, err
	}

	entries, err := ts.blobStore.driver.List(ctx, pathSpec)
	if err != nil {
		switch err := err.(type) {
		case storagedriver.PathNotFoundError:
			return 0, distribution.ErrRepositoryUnknown{Name: ts.repository.Name()}
		default:
			return 0, err
		}
	}

	sort.Strings(entries)
	var tagNames []string
	for _, entry := range entries {
		if len(tagNames) == len(tags) {
			copy(tags, tagNames)
			return len(tags), nil
		}

		_, filename := path.Split(entry)
		if filename > last {
			tagNames = append(tagNames, filename)
		}
	}

	n = copy(tags, tagNames)
	return n, io.EOF
}

// exists returns true if the specified manifest tag exists in the repository.
func (ts *tagStore) exists(ctx context.Context, tag string) (bool, error) {
	tagPath, err := pathFor(manifestTagCurrentPathSpec{
		name: ts.repository.Name(),
		tag:  tag,
	})

	if err != nil {
		return false, err
	}

	exists, err := exists(ctx, ts.blobStore.driver, tagPath)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Tag tags the digest with the given tag, updating the the store to point at
// the current tag. The digest must point to a manifest.
func (ts *tagStore) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	currentPath, err := pathFor(manifestTagCurrentPathSpec{
		name: ts.repository.Name(),
		tag:  tag,
	})

	if err != nil {
		return err
	}

	lbs := ts.linkedBlobStore(ctx, tag)

	// Link into the index
	if err := lbs.linkBlob(ctx, desc); err != nil {
		return err
	}

	// Overwrite the current link
	return ts.blobStore.link(ctx, currentPath, desc.Digest)
}

// resolve the current revision for name and tag.
func (ts *tagStore) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	currentPath, err := pathFor(manifestTagCurrentPathSpec{
		name: ts.repository.Name(),
		tag:  tag,
	})

	if err != nil {
		return distribution.Descriptor{}, err
	}

	revision, err := ts.blobStore.readlink(ctx, currentPath)
	if err != nil {
		switch err.(type) {
		case storagedriver.PathNotFoundError:
			return distribution.Descriptor{}, distribution.ErrTagUnknown{Tag: tag}
		}

		return distribution.Descriptor{}, err
	}

	return distribution.Descriptor{Digest: revision}, nil
}

// delete removes the tag from repository, including the history of all
// revisions that have the specified tag.
func (ts *tagStore) Untag(ctx context.Context, tag string) error {
	tagPath, err := pathFor(manifestTagPathSpec{
		name: ts.repository.Name(),
		tag:  tag,
	})

	if err != nil {
		return err
	}

	return ts.blobStore.driver.Delete(ctx, tagPath)
}

// linkedBlobStore returns the linkedBlobStore for the named tag, allowing one
// to index manifest blobs by tag name. While the tag store doesn't map
// precisely to the linked blob store, using this ensures the links are
// managed via the same code path.
func (ts *tagStore) linkedBlobStore(ctx context.Context, tag string) *linkedBlobStore {
	return &linkedBlobStore{
		blobStore:  ts.blobStore,
		repository: ts.repository,
		ctx:        ctx,
		linkPathFns: []linkPathFunc{func(name string, dgst digest.Digest) (string, error) {
			return pathFor(manifestTagIndexEntryLinkPathSpec{
				name:     name,
				tag:      tag,
				revision: dgst,
			})

		}},
	}
}

// Lookup recovers a list of tags which refer to this digest.  When a manifest is deleted by
// digest, tag entries which point to it need to be recovered to avoid dangling tags.
func (ts *tagStore) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	// An efficient implementation of this will require changes to the S3 driver.
	return make([]string, 0), nil
}
