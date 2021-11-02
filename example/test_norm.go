package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	for i := 0; i < 20; i++ {
		fmt.Printf("%v\n", rand.NormFloat64()*10+5)
	}
}
