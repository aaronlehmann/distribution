package proxy

import (
	"sync"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
)

type mockTagStore struct {
	mapping map[string]distribution.Descriptor
	sync.Mutex
}

var _ distribution.TagService = &mockTagStore{}

func (m *mockTagStore) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	m.Lock()
	defer m.Unlock()

	if d, ok := m.mapping[tag]; ok {
		return d, nil
	}
	return distribution.Descriptor{}, distribution.ErrTagUnknown{}
}

func (m *mockTagStore) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	m.Lock()
	defer m.Unlock()

	m.mapping[tag] = desc
	return nil
}

func (m *mockTagStore) Untag(ctx context.Context, tag string) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.mapping[tag]; ok {
		delete(m.mapping, tag)
		return nil
	}
	return distribution.ErrTagUnknown{}
}

func (m *mockTagStore) Enumerate(ctx context.Context, refs []string, last string) (n int, err error) {
	panic("not implemented")
}

func (m *mockTagStore) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	panic("not implemented")
}

func testProxyTagService(local, remote map[string]distribution.Descriptor) *proxyTagService {
	if local == nil {
		local = make(map[string]distribution.Descriptor)
	}
	if remote == nil {
		remote = make(map[string]distribution.Descriptor)
	}
	return &proxyTagService{
		localTags:  &mockTagStore{mapping: local},
		remoteTags: &mockTagStore{mapping: remote},
	}
}

func TestGet(t *testing.T) {
	remoteDesc := distribution.Descriptor{Size: 42}
	remoteTag := "remote"
	proxyTags := testProxyTagService(map[string]distribution.Descriptor{remoteTag: remoteDesc}, nil)

	ctx := context.Background()

	d, err := proxyTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatal(err)
	}

	if d != remoteDesc {
		t.Fatal("unable to get put tag")
	}

	local, err := proxyTags.localTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatal("remote tag not pulled into store")
	}

	if local != remoteDesc {
		t.Fatalf("unexpected descriptor pulled through")
	}

	// Manually overwrite remote tag
	newRemoteDesc := distribution.Descriptor{Size: 43}
	err = proxyTags.remoteTags.Tag(ctx, remoteTag, newRemoteDesc)
	if err != nil {
		t.Fatal(err)
	}

	d, err = proxyTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatal(err)
	}

	if d != newRemoteDesc {
		t.Fatal("unable to get put tag")
	}

	_, err = proxyTags.localTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatal("remote tag not pulled into store")
	}

	// untag, ensure it's removed locally, but present in remote
	err = proxyTags.Untag(ctx, remoteTag)
	if err != nil {
		t.Fatal(err)
	}

	_, err = proxyTags.localTags.Get(ctx, remoteTag)
	if err == nil {
		t.Fatalf("Expected error getting Untag'd tag")
	}

	_, err = proxyTags.remoteTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatalf("remote tag should not be untagged with proxyTag.Untag")
	}

	_, err = proxyTags.Get(ctx, remoteTag)
	if err != nil {
		t.Fatal("untagged tag should be pulled through")
	}

}
