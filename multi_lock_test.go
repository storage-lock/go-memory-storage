package memory_storage

import (
	"context"
	"errors"
	"testing"

	storage_lock "github.com/storage-lock/go-storage-lock"
)

func TestAtomicMultiLockAllOrNothing(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	first, err := storage_lock.NewStorageLock(store, "b")
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Lock(ctx, "other"); err != nil {
		t.Fatal(err)
	}
	left, err := storage_lock.NewStorageLock(store, "a")
	if err != nil {
		t.Fatal(err)
	}
	right, err := storage_lock.NewStorageLock(store, "b")
	if err != nil {
		t.Fatal(err)
	}
	multi, err := storage_lock.NewMultiLock("me", right, left)
	if err != nil {
		t.Fatal(err)
	}
	if ids := multi.LockIds(); len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("locks were not ordered: %v", ids)
	}
	err = multi.Lock(ctx)
	if !errors.Is(err, storage_lock.ErrLockBusy) {
		t.Fatalf("expected busy, got %v", err)
	}
	if _, err := store.Get(ctx, "a"); !errors.Is(err, storage_lock.ErrLockNotFound) {
		t.Fatalf("rejected batch still created a: %v", err)
	}
	if err := first.UnLock(ctx, "other"); err != nil {
		t.Fatal(err)
	}
	if err := multi.Lock(ctx); err != nil {
		t.Fatal(err)
	}
	if multi.HeldCount() != 2 {
		t.Fatalf("held %d", multi.HeldCount())
	}
	again, err := storage_lock.NewStorageLock(store, "a")
	if err != nil {
		t.Fatal(err)
	}
	if err := again.TryLock(ctx, "third"); !errors.Is(err, storage_lock.ErrLockBusy) {
		t.Fatalf("sibling lock was not held: %v", err)
	}
	if err := multi.UnLock(ctx); err != nil {
		t.Fatal(err)
	}
	if multi.HeldCount() != 0 {
		t.Fatalf("still held %d", multi.HeldCount())
	}
	if err := again.TryLock(ctx, "third"); err != nil {
		t.Fatalf("lock stayed busy after atomic release: %v", err)
	}
}
