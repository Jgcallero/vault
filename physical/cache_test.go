package physical

import (
	"testing"

	"github.com/hashicorp/vault/helper/logformat"
	log "github.com/jefferai/logxi/v1"
)

func TestCache(t *testing.T) {
	logger := logformat.NewVaultLogger(log.LevelTrace)

	inm := NewInmem(logger)
	cache := NewCache(inm, 0)
	testBackend(t, cache)
	testBackend_ListPrefix(t, cache)
}

func TestCache_Purge(t *testing.T) {
	logger := logformat.NewVaultLogger(log.LevelTrace)

	inm := NewInmem(logger)
	cache := NewCache(inm, 0)

	ent := &Entry{
		Key:   "foo",
		Value: []byte("bar"),
	}
	err := cache.Put(ent)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete from under
	inm.Delete("foo")

	// Read should work
	out, err := cache.Get("foo")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out == nil {
		t.Fatalf("should have key")
	}

	// Clear the cache
	cache.Purge()

	// Read should fail
	out, err = cache.Get("foo")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out != nil {
		t.Fatalf("should not have key")
	}
}
