package main

import (
	"context"
	"fmt"
	"time"
)

func sampleOperation(ctx context.Context,
	msg string, msDelay time.Duration) <-chan string {
	out := make(chan string)

	go func() {
		for {
			select {
			case <-time.After(msDelay * time.Millisecond):
				out <- fmt.Sprintf("%v operation completed", msg)
				return
			case <-ctx.Done():
				out <- fmt.Sprintf("%v aborted", msg)
				return
			}
		}
	}()
	return out
}
func main() {
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)

	webserver := sampleOperation(ctx, "webserver", 200)
	microservice := sampleOperation(ctx, "microservice", 500)
	database := sampleOperation(ctx, "database", 900)

Mainloop:
	for {
		select {
		case val := <-webserver:
			fmt.Println(":", val)
		case val := <-microservice:
			fmt.Println(":", val)
			fmt.Println("cancel context:")
			cancelCtx()
			break Mainloop
		case val := <-database:
			fmt.Println(":", val)
		}
	}
	fmt.Println(":", <-database)
}
