aPLib implementation in Go
==========================

Usage of the library should be pretty simple.

```bash
go get github.com/hatching/aplib
```

```go
import (
    "github.com/hatching/aplib"
)

func hello() {
    v1 := aplib.Compress([]byte("hello world"))
    v2 := aplib.Decompress(v1)
    // ...
}
```
