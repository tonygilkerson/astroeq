package main

import (
	"fmt"
	"github.com/tonygilkerson/astroeq/pkg/hid"
)

func main() {
	fmt.Println("handset")
	foo := hid.Foo()
	fmt.Println(foo)
}
