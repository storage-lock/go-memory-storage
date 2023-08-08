package main

import (
	"fmt"
	memory_storage "github.com/storage-lock/go-memory-storage"
)

func main() {

	// 直接创建就可以使用了
	storage := memory_storage.NewMemoryStorage()
	fmt.Println(storage.GetName())

}
