# Memory Storage

# 一、这是什么？

基于内存实现的Storage，相当于是把锁存放在内存中。

# 二、安装

```bash
go get -u github.com/storage-lock/go-memory-storage
```

# 三、API示例

```go
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
```



