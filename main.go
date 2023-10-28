package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Define flags
	url := flag.String("u", "", "URL to request")
	n := flag.Int("n", 1, "Number of requests to make")
	c := flag.Int("c", 1, "Concurrency")
	flag.Parse()

	// Validate flags
	if *url == "" {
		fmt.Println("Please provide a URL with the -u flag.")
		return
	}

	codeCounts := make(map[int]int)
	var codeMutex sync.Mutex

	startTime := time.Now()

	makeRequest := func(startSignal <-chan struct{}) {
		<-startSignal // Wait for the signal to start the request
		resp, err := http.Get(*url)
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer resp.Body.Close()

		codeMutex.Lock()
		codeCounts[resp.StatusCode]++
		codeMutex.Unlock()
	}

	for i := 0; i < *n; i += *c {
		var wg sync.WaitGroup
		batchSize := min(*c, *n-i)
		wg.Add(batchSize)

		startSignal := make(chan struct{})
		for j := 0; j < batchSize; j++ {
			go func() {
				defer wg.Done()
				makeRequest(startSignal)
			}()
		}

		close(startSignal) // Send signal to start all requests in this batch
		wg.Wait()
		fmt.Printf("Finished %d requests\n", min(i+batchSize, *n))
	}

	totalTime := time.Since(startTime)
	meanTime := totalTime / time.Duration(*n)

	fmt.Println("Total time for all requests:", totalTime)
	fmt.Println("Mean time for all requests:", meanTime)
	fmt.Println("Number of responses by HTTP code:")
	for code, count := range codeCounts {
		fmt.Printf("HTTP %d: %d\n", code, count)
	}
}

// min is used to handle the case where the total number of requests is not
// evenly divisible by the concurrency level.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
