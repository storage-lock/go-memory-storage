package memory_storage

import (
	"context"
	"errors"
	"testing"

	storage_lock "github.com/storage-lock/go-storage-lock"
)

func TestLockMetricsCountAcquireAndBusy(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStorage()
	metrics := storage_lock.NewLockMetrics()
	options := storage_lock.NewStorageLockOptionsWithLockId("metrics-lock")
	options.EventListeners = append(options.EventListeners, metrics.Listener())
	options.DisableWatchDog = true
	first, err := storage_lock.NewStorageLockWithOptions(store, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := storage_lock.NewStorageLockWithOptions(store, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.TryLock(ctx, "holder"); err != nil {
		t.Fatal(err)
	}
	if err := second.TryLock(ctx, "other"); !errors.Is(err, storage_lock.ErrLockBusy) {
		t.Fatalf("expected busy, got %v", err)
	}
	snapshot := metrics.Snapshot()
	if snapshot.Acquired != 1 || snapshot.Busy == 0 {
		t.Fatalf("metrics = %+v", snapshot)
	}
}
