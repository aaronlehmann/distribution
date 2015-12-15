package storage

import (
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"io"
)

type tagsTestEnv struct {
	ts  distribution.TagService
	ctx context.Context
}

func testTagStore(t *testing.T) *tagsTestEnv {
	ctx := context.Background()
	d := inmemory.New()
	reg, err := NewRegistry(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	repoRef, _ := reference.ParseNamed("a/b")
	repo, err := reg.Repository(ctx, repoRef)
	if err != nil {
		t.Fatal(err)
	}

	return &tagsTestEnv{
		ctx: ctx,
		ts:  repo.Tags(ctx),
	}
}

func TestTagStoreTag(t *testing.T) {
	env := testTagStore(t)
	tags := env.ts
	ctx := env.ctx

	d := distribution.Descriptor{}
	err := tags.Tag(ctx, "latest", d)
	if err == nil {
		t.Errorf("unexpected error putting malformed descriptor : %s", err)
	}

	d.Digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	err = tags.Tag(ctx, "latest", d)
	if err != nil {
		t.Error(err)
	}

	d1, err := tags.Get(ctx, "latest")
	if err != nil {
		t.Error(err)
	}

	if d1.Digest != d.Digest {
		t.Error("put and get digest differ")
	}

	// Overwrite existing
	d.Digest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	err = tags.Tag(ctx, "latest", d)
	if err != nil {
		t.Error(err)
	}

	d1, err = tags.Get(ctx, "latest")
	if err != nil {
		t.Error(err)
	}

	if d1.Digest != d.Digest {
		t.Error("put and get digest differ")
	}
}

func TestTagStoreUnTag(t *testing.T) {
	env := testTagStore(t)
	tags := env.ts
	ctx := env.ctx
	desc := distribution.Descriptor{Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}

	err := tags.Untag(ctx, "latest")
	if err == nil {
		t.Errorf("Expected error untagging non-existant tag")
	}

	err = tags.Tag(ctx, "latest", desc)
	if err != nil {
		t.Error(err)
	}

	err = tags.Untag(ctx, "latest")
	if err != nil {
		t.Error(err)
	}

	_, err = tags.Get(ctx, "latest")
	if err == nil {
		t.Error("Expected error getting untagged tag")
	}
}

func TestTagEnumerate(t *testing.T) {
	env := testTagStore(t)
	tags := env.ts
	ctx := env.ctx
	alpha := "abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < len(alpha); i++ {
		tag := alpha[i]
		desc := distribution.Descriptor{Digest: "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}
		err := tags.Tag(ctx, string(tag), desc)
		if err != nil {
			t.Error(err)
		}
	}

	// collect all in one go
	all := make([]string, len(alpha), len(alpha))
	n, err := tags.Enumerate(ctx, all, "")
	if err != io.EOF {
		t.Error(err)
	}
	if n != len(alpha) {
		t.Errorf("Unexpected count returned from enumerate")
	}

	for i, c := range all {
		if c != string(alpha[i]) {
			t.Errorf("unexpected tag in enumerate %s", c)
		}
	}

	some := make([]string, 5)
	n, err = tags.Enumerate(ctx, some, "")
	if err != nil {
		t.Errorf("Unexepected error enumerating tags")
	}

	for i, c := range some {
		if c != string(alpha[i]) {
			t.Errorf("unexpected tag in enumerate %s", c)
		}
	}

	more := make([]string, 5)
	n, err = tags.Enumerate(ctx, more, "")
	if err != nil {
		t.Errorf("Unexepected error enumerating tags")
	}
	if n != len(more) {
		t.Errorf("Unexpected enumerate count")
	}
	for i, c := range more {
		if c != string(alpha[i]) {
			t.Errorf("unexpected tag in enumerate %s", c)
		}
	}

	last := more[len(more)-1]
	more = make([]string, 5)
	n, err = tags.Enumerate(ctx, more, last)
	if err != nil {
		t.Errorf("Unexepected error enumerating tags")
	}
	if n != len(more) {
		t.Errorf("Unexpected enumerate count")
	}
	for i, c := range more {
		if c != string(alpha[i+len(more)]) {
			t.Errorf("unexpected tag in enumerate %s", c)
		}
	}
}
