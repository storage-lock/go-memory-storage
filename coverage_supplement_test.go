package memory_storage

import (
	"context"
	"testing"
	"time"

	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
)

// 覆盖 Capabilities / Watch / CRUD 错误分支 / List 损坏数据分支

func newTestLockInfo() *storage.LockInformation {
	return &storage.LockInformation{
		OwnerId:         "owner-cov",
		Version:         1,
		LockCount:       1,
		LockBeginTime:   time.Now(),
		LeaseExpireTime: time.Now().Add(time.Minute),
	}
}

func TestCoverageCapabilities(t *testing.T) {
	s := NewMemoryStorage()
	caps := s.Capabilities()
	assert.Contains(t, caps, storage.CapabilityCAS)
	assert.Contains(t, caps, storage.CapabilityReliableTime)
	assert.Contains(t, caps, storage.CapabilityAtomicDelete)
	assert.Contains(t, caps, storage.CapabilityWatch)
}

func TestCoverageWatch(t *testing.T) {
	s := NewMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := s.Watch(ctx, "lock-watch-cov")
	assert.Nil(t, err)
	assert.NotNil(t, ch)
	// Watch 空 lockId 直接返回错误
	_, err = s.Watch(ctx, "")
	assert.NotNil(t, err)
}

func TestCoverageUpdateNotExists(t *testing.T) {
	s := NewMemoryStorage()
	err := s.UpdateWithVersion(context.Background(), "lock-missing", 1, 2, newTestLockInfo())
	assert.Equal(t, storage_lock.ErrVersionMiss, err)
}

func TestCoverageUpdateVersionMismatch(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()
	assert.Nil(t, s.CreateWithVersion(ctx, "lock-mismatch", 1, newTestLockInfo()))
	err := s.UpdateWithVersion(ctx, "lock-mismatch", 999, 2, newTestLockInfo())
	assert.Equal(t, storage_lock.ErrVersionMiss, err)
}

func TestCoverageCreateAlreadyExists(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()
	assert.Nil(t, s.CreateWithVersion(ctx, "lock-dup", 1, newTestLockInfo()))
	err := s.CreateWithVersion(ctx, "lock-dup", 1, newTestLockInfo())
	assert.Equal(t, storage_lock.ErrLockAlreadyExists, err)
}

func TestCoverageDeleteNotExists(t *testing.T) {
	s := NewMemoryStorage()
	err := s.DeleteWithVersion(context.Background(), "lock-missing", 1, newTestLockInfo())
	assert.Equal(t, storage_lock.ErrLockNotFound, err)
}

func TestCoverageDeleteVersionMismatch(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()
	assert.Nil(t, s.CreateWithVersion(ctx, "lock-delmismatch", 1, newTestLockInfo()))
	err := s.DeleteWithVersion(ctx, "lock-delmismatch", 999, newTestLockInfo())
	assert.Equal(t, storage_lock.ErrVersionMiss, err)
}

func TestCoverageListCorruptedEntry(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()
	// 直接注入一条损坏的 JSON，覆盖 List 的 unmarshal 错误分支
	s.storageLock.Lock()
	s.storageMap["lock-corrupted"] = &MemoryStorageValue{
		LockId:                    "lock-corrupted",
		Version:                   1,
		LockInformationJsonString: "{invalid-json",
	}
	s.storageLock.Unlock()
	_, err := s.List(ctx)
	assert.NotNil(t, err)
}
