package main

import (
	"fmt"

	"golang.org/x/example/stringutil"
	"github.com/tonygilkerson/astroeq/de/foo"
)

var Foo int

func main() {
	fmt.Println(stringutil.Reverse("I am DE"))
	var t foo.Thing
	fmt.Printf("%T\n",t)
}
