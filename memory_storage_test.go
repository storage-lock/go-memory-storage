package memory_storage

import (
	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"testing"
)

func TestMemoryStorage(t *testing.T) {
	storage_test_helper.TestStorage(t, NewMemoryStorage())
}
