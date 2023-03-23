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

workersWg

}
