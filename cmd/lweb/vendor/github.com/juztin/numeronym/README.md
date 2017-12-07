# Numeronym

This is just a simple library that creates numeronyms *(i18n => internationalization)*.

To install just do a simple `go get github.com/juztin/numeronym`  
To use:

```go
package main

import (
	"fmt"

	"github.com/juztin/numeronym"
)

func main() {
	text := []byte("internationalization")
	b := numeronym.Parse(text)
	fmt.Println(string(b))

	// OUTPUTS:
	//	i18n
}
```
