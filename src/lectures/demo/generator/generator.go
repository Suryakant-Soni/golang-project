package main

import (
	"fmt"
	"math/rand"
	"time"
)

func generateRandInt(min, max int) <-chan int {
	out := make(chan int, 3)
	go func() {
		for {
			out <- rand.Intn(max-min) + min
		}
	}()
	return out
}
// func ()  {
	
// }
func main() {
	rand.Seed(time.Now().UnixNano())
	randInt := generateRandInt(1, 100)
	fmt.Println("generate Rand Int")
	fmt.Println(":", <-randInt)
	fmt.Println(":", <-randInt)
	fmt.Println(":", <-randInt)
	fmt.Println(":", <-randInt)
}
