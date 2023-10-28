package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Requester is repsonsible for making a single HTTP request
type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

// App represents the individual components of the tester
type App struct {
	URL         string
	NumRequests int
	Concurrency int
	Requester   Requester
	Out         io.Writer
}

// Run will execute a number of concurrent requests to a URL
func (a *App) Run() {
	codeCounts := make(map[int]int)
	var codeMutex sync.Mutex

	startTime := time.Now()

	makeRequest := func(startSignal <-chan struct{}) {
		<-startSignal
		req, err := http.NewRequest("GET", a.URL, nil)
		if err != nil {
			fmt.Fprintln(a.Out, "Error creating request:", err)
			return
		}
		resp, err := a.Requester.Do(req)
		if err != nil {
			fmt.Fprintln(a.Out, "Error making request:", err)
			return
		}
		defer resp.Body.Close()

		codeMutex.Lock()
		codeCounts[resp.StatusCode]++
		codeMutex.Unlock()
	}

	for i := 0; i < a.NumRequests; i += a.Concurrency {
		var wg sync.WaitGroup
		batchSize := min(a.Concurrency, a.NumRequests-i)
		wg.Add(batchSize)

		startSignal := make(chan struct{})
		for j := 0; j < batchSize; j++ {
			go func() {
				defer wg.Done()
				makeRequest(startSignal)
			}()
		}

		close(startSignal)
		wg.Wait()
		fmt.Fprintf(a.Out, "Finished %d requests\n", min(i+batchSize, a.NumRequests))
	}

	totalTime := time.Since(startTime)
	meanTime := totalTime / time.Duration(a.NumRequests)

	fmt.Fprintln(a.Out, "Total time for all requests:", totalTime)
	fmt.Fprintln(a.Out, "Mean time for all requests:", meanTime)
	fmt.Fprintln(a.Out, "Number of responses by HTTP code:")
	for code, count := range codeCounts {
		fmt.Fprintf(a.Out, "HTTP %d: %d\n", code, count)
	}
}

func main() {
	url := flag.String("u", "", "URL to request")
	n := flag.Int("n", 1, "Number of requests to make")
	c := flag.Int("c", 1, "Concurrency")
	flag.Parse()

	if *url == "" {
		fmt.Println("Please provide a URL with the -u flag.")
		return
	}

	app := &App{
		URL:         *url,
		NumRequests: *n,
		Concurrency: *c,
		Requester:   http.DefaultClient,
		Out:         os.Stdout,
	}
	app.Run()
}

// min is used to handle the case where the total number of requests is not
// evenly divisible by the concurrency level.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
