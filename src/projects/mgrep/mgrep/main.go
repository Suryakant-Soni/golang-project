package main

import (
	"fmt"
	"mgrep/worker"
	"mgrep/worklist"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexflint/go-arg"
)

func discoverDirs(wl *worklist.Worklist, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			nextPath := filepath.Join(path, entry.Name())
			// there is further directory
			discoverDirs(wl, nextPath)
		} else {
			// add a new job for the file found
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}

}

var args struct {
	SearchTerm string `arg:"positional,required"`
	SearchDir  string `arg:"positional"`
}

func main() {
	// parsing is mandatory
	arg.MustParse(&args)

	var workersWg sync.WaitGroup
	// getting a new worklist with buffer of 100
	wl := worklist.New(100)
	// this is  new channel where workers dump the result
	results := make(chan worker.Result, 100)
	numWorkers := 10

	workersWg.Add(1)
	go func() {
		defer workersWg.Done()
		// this will get all directories and send them in the jobs channel
		discoverDirs(&wl, args.SearchDir)
		wl.Finalize(numWorkers)
	}()
	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		// this goroutine will find the searchterm in the file and send it to resutls chananel
		go func() {
			defer workersWg.Done()
			for {
				workEntry := wl.Next()
				if workEntry.Path != "" {
					workerResult := worker.FindInFile(workEntry.Path, args.SearchTerm)
					if workerResult != nil {
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				} else {
					return
				}
			}
		}()
	}
	blockWorkersWg := make(chan struct{})
	go func() {
		// create another thread to block as the main thread will start to print the ouput as and when received
		workersWg.Wait()
		close(blockWorkersWg)
	}()

	// create a wait group which will hold the main routine
	var displayWg sync.WaitGroup

	displayWg.Add(1)
	go func() {
		for {
			select {
			case r := <-results:
				fmt.Printf("%v - %v  - %v \n", r.Path, r.LineNum, r.Line)
			case <-blockWorkersWg:
				// we need to check here additionally that results is not left with any more entry and everything is printed
				if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()
	displayWg.Wait()
}
