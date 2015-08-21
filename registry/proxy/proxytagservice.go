package proxy

import (
	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
)

// proxyTagService supports local and remote lookup of tags.
type proxyTagService struct {
	localTags  distribution.TagService
	remoteTags distribution.TagService
}

var _ distribution.TagService = proxyTagService{}

// Get attempts to get the most recent digest for the tag by checking the remote
// tag service first and then caching it locally.  If the remote is unavailable
// the local association is returned
func (pt proxyTagService) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	desc, err := pt.remoteTags.Get(ctx, tag)
	if err == nil {
		err := pt.localTags.Tag(ctx, tag, desc)
		if err != nil {
			return distribution.Descriptor{}, err
		}
		return desc, nil
	}

	desc, err = pt.localTags.Get(ctx, tag)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	return desc, nil
}

func (pt proxyTagService) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	return distribution.ErrUnsupported
}

func (pt proxyTagService) Untag(ctx context.Context, tag string) error {
	err := pt.localTags.Untag(ctx, tag)
	if err != nil {
		return err
	}
	return nil
}

func (pt proxyTagService) Enumerate(ctx context.Context, refs []string, last string) (n int, err error) {
	return 0, distribution.ErrUnsupported
}

func (pt proxyTagService) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	return []string{}, distribution.ErrUnsupported
}
