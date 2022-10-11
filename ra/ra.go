package main

import (
	"fmt"

	"github.com/tonygilkerson/astroeq/de/foo"
	"golang.org/x/example/stringutil"
)

func main() {

	fmt.Println(stringutil.Reverse("I am RA"))
	var t foo.Thing
	fmt.Printf("%T\n", t)
	fmt.Println(foo.Hi())
	fmt.Println(foo.HiHo())
	fmt.Println(foo.HiHoa())
}
