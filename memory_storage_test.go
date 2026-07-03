package memory_storage

import (
	"context"
	storage_pkg "github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"sync"
	"testing"
	"time"
)

func TestMemoryStorage(t *testing.T) {
	storage_test_helper.TestStorage(t, NewMemoryStorage())
}

// TestMemoryStorageListConcurrentNoCrash 钉死 List 无锁漏洞：
// 修复前 List 遍历 storageMap 不持锁，与并发 Create/Update/Delete（持写锁）触发
// Go 运行时 fatal error: concurrent map read and map write（不可 recover，进程崩溃）。
// 修复后 List 持读锁拷贝，并发安全。
func TestMemoryStorageListConcurrentNoCrash(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()

	// 预置一把锁
	info := &storage_pkg.LockInformation{
		OwnerId:         "owner-list-test",
		Version:         1,
		LockCount:       1,
		LockBeginTime:   time.Now(),
		LeaseExpireTime: time.Now().Add(time.Minute),
	}
	if err := s.CreateWithVersion(ctx, "lock-list-test", 1, info); err != nil {
		t.Fatalf("预置锁失败: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// 并发 List
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_, _ = s.List(ctx)
			}
		}()
	}

	// 并发 Create/Update/Delete 同一锁，制造 map 写
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				lockId := "lock-list-rw-" + string(rune('a'+g))
				_ = s.CreateWithVersion(ctx, lockId, 1, info)
				_ = s.UpdateWithVersion(ctx, lockId, 1, 2, info)
				_ = s.DeleteWithVersion(ctx, lockId, 2, info)
			}
		}(i)
	}

	// 让 List 跑一会
	time.Sleep(time.Millisecond * 200)
	close(stop)
	wg.Wait()
	// 走到这里说明没有 fatal crash（List 无锁漏洞修复有效）

	_ = storage_lock.ErrVersionMiss // 防止未使用告警
}

