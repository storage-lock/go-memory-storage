package memory_storage

import (
	"context"
	"errors"
	"testing"

	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

func TestCompareAndSwapBatchAllOrNothing(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	existing := &storage.LockInformation{LockId: "b", Version: 3, LockCount: 1, OwnerId: "other"}
	if err := store.CreateWithVersion(ctx, "b", 3, existing); err != nil {
		t.Fatal(err)
	}

	created := &storage.LockInformation{LockId: "a", Version: 1, LockCount: 1, OwnerId: "me"}
	updated := &storage.LockInformation{LockId: "b", Version: 4, LockCount: 1, OwnerId: "me"}
	err := store.CompareAndSwapBatch(ctx, []storage.CASOp{
		{LockID: "a", NewVersion: 1, Information: created},
		{LockID: "b", ExpectedVersion: 9, NewVersion: 4, Information: updated},
	})
	if !errors.Is(err, storage_lock.ErrVersionMiss) {
		t.Fatalf("expected version miss, got %v", err)
	}
	if _, err := store.Get(ctx, "a"); !errors.Is(err, storage_lock.ErrLockNotFound) {
		t.Fatalf("failed batch created key a: %v", err)
	}
	raw, err := store.Get(ctx, "b")
	if err != nil {
		t.Fatal(err)
	}
	if raw != existing.ToJsonString() {
		t.Fatalf("key b changed on failed batch: %s", raw)
	}

	if err := store.CompareAndSwapBatch(ctx, []storage.CASOp{
		{LockID: "a", NewVersion: 1, Information: created},
		{LockID: "b", ExpectedVersion: 3, NewVersion: 4, Information: updated},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	raw, err = store.Get(ctx, "b")
	if err != nil || raw != updated.ToJsonString() {
		t.Fatalf("key b = %s, err %v", raw, err)
	}
}
