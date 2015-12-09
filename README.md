# TTL Slice Map for Caching

## How To Use

```sh
go get "github.com/yudai/ttlslicemap"
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/yudai/ttlslicemap"
)

func main() {
	// 5 minutes TTL
	tsm := ttlslicemap.New(time.Minute * 5)

	// Add items
	tsm.Add("one", "foo")
	tsm.Add("one", "bar")
	items, exists := tsm.Get("one")
	fmt.Println(items)  // => ["foo", "bar"]
	fmt.Println(exists) // => true

	// Add more items
	tsm.Add("two", "red")
	tsm.Add("three", "blue")
	fmt.Println(tsm.Count()) // => 3

	// Then remove one
	tsm.Remove("one")
	fmt.Println(tsm.Count()) // => 2

	// Wait expire
	time.Sleep(time.Minute * 6)
	fmt.Println(tsm.Count()) // => 0
}
```

## License
The MIT License
