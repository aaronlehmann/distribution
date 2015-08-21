package distribution

import (
	"github.com/docker/distribution/context"
)

// TagService provides access to information about tagged objects.
type TagService interface {
	// Get retrieves the descriptor identified by the tag. Some
	// implementations may differentiate between "trusted" tags and
	// "untrusted" tags. If a tag is "untrusted", the mapping will be returned
	// as an ErrTagUntrusted error, with the target descriptor.
	Get(ctx context.Context, tag string) (Descriptor, error)

	// Tag associates the tag with the provided descriptor, updating the
	// current association, if needed.
	Tag(ctx context.Context, tag string, desc Descriptor) error

	// Untag removes the given tag association
	Untag(ctx context.Context, tag string) error

	// Enumerate fills 'refs' with a lexigraphically sorted set of tags up to
	// the size of 'refs' and returns 'n' for the number of entries which were
	// filled.  'last' contains an offset in the tag set and can be used to resume
	// iteration.
	Enumerate(ctx context.Context, refs []string, last string) (n int, err error)

	// Lookup returns the set of tags referencing the given digest
	Lookup(ctx context.Context, digest Descriptor) ([]string, error)
}
