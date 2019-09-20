// Experimented managing go-routines in recursion and take data in channels.

package main

import (
	"fmt"
	"sync"
	"time"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

func dec_routine() {
	lock.Lock()
	r_num = r_num - 1
	//fmt.Println("Exit Crawler Routine : ", r_num)
	wg.Done()
	lock.Unlock()
}

func inc_routine() {
	lock.Lock()
	r_num = r_num + 1
	//fmt.Println("New Crawler Routine : ", r_num)
	wg.Add(1)
	lock.Unlock()
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, chan_for_craw chan string) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	//defer wg.Done()
	defer dec_routine()
	if _, found := visited[url]; !found {
		lock.Lock()
		visited[url] = true
		lock.Unlock()
		if depth <= 0 {
			return
		}
		chan_for_craw <- url
		body, urls, err := fetcher.Fetch(url)
		body = body + ""

		if err != nil {
			//fmt.Println("Error: ", err)
			return
		}
		//fmt.Println("Found: ", url, body)

		for _, u := range urls {
			inc_routine()
			go Crawl(u, depth-1, fetcher, chan_for_craw)
		}
		return
	} else {
		fmt.Println("Skipping ... ", url)
		return
	}
}

var visited map[string]bool
var lock = sync.RWMutex{}
var r_num int
var wg sync.WaitGroup

func main() {
	visited = make(map[string]bool, 1000)
	chan_for_craw := make(chan string)
	r_num = 0
	lock.Lock()
	wg.Add(1)
	r_num += 1
	lock.Unlock()
	//fmt.Println("New Crawler Routine: ", r_num)
	fmt.Println("Starting from https://golang.org/")
	go Crawl("https://golang.org/", 4, fetcher, chan_for_craw)
	go func() {
		//fmt.Println("Closes Channel Routine: Will close channel after all routines finish... waiting")
		wg.Wait()
		close(chan_for_craw)
		//fmt.Println("Closes Channel Routine: Channel closed")
	}()

	go func() {
		for i := range chan_for_craw {
			fmt.Printf("Printing Routine: Found: %s\n", i)
		}
	}()

	//fmt.Println("Main: Wait for go routines to finish")
	wg.Wait()
	//fmt.Println("Main: All go routines finished")

}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	//fmt.Printf("Fetching: %s ...\n", url)
	time.Sleep(1 * time.Second)
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

/*
Fetching: https://golang.org/ ...
found: https://golang.org/ "The Go Programming Language"
Fetching: https://golang.org/pkg/ ...
found: https://golang.org/pkg/ "Packages"
Fetching: https://golang.org/ ...
found: https://golang.org/ "The Go Programming Language"
Fetching: https://golang.org/pkg/ ...
found: https://golang.org/pkg/ "Packages"
Fetching: https://golang.org/cmd/ ...
not found: https://golang.org/cmd/
Fetching: https://golang.org/cmd/ ...
not found: https://golang.org/cmd/
Fetching: https://golang.org/pkg/fmt/ ...
found: https://golang.org/pkg/fmt/ "Package fmt"
Fetching: https://golang.org/ ...
found: https://golang.org/ "The Go Programming Language"
Fetching: https://golang.org/pkg/ ...
found: https://golang.org/pkg/ "Packages"
Fetching: https://golang.org/pkg/os/ ...
found: https://golang.org/pkg/os/ "Package os"
Fetching: https://golang.org/ ...
found: https://golang.org/ "The Go Programming Language"
Fetching: https://golang.org/pkg/ ...
found: https://golang.org/pkg/ "Packages"
Fetching: https://golang.org/cmd/ ...
not found: https://golang.org/cmd/
*/

/*
Fetching: https://golang.org/ ...
found: https://golang.org/ "The Go Programming Language"
Fetching: https://golang.org/pkg/ ...
found: https://golang.org/pkg/ "Packages"
Skipping ...  https://golang.org/
Fetching: https://golang.org/cmd/ ...
not found: https://golang.org/cmd/
Fetching: https://golang.org/pkg/fmt/ ...
found: https://golang.org/pkg/fmt/ "Package fmt"
Skipping ...  https://golang.org/
Skipping ...  https://golang.org/pkg/
Fetching: https://golang.org/pkg/os/ ...
found: https://golang.org/pkg/os/ "Package os"
Skipping ...  https://golang.org/
Skipping ...  https://golang.org/pkg/
Skipping ...  https://golang.org/cmd/
*/

/*
Crawling..https://golang.org/
Fetching: https://golang.org/ ...
Found:  https://golang.org/ The Go Programming Language
Crawling sub.. https://golang.org/pkg/
Crawling sub.. https://golang.org/cmd/
Fetching: https://golang.org/cmd/ ...
Master Status +1 :  https://golang.org/
Fetching: https://golang.org/pkg/ ...
Found:  https://golang.org/pkg/ Packages
Crawling sub.. https://golang.org/
Error:  not found: https://golang.org/cmd/
Crawling sub.. https://golang.org/cmd/
Master Status +1 :  https://golang.org/pkg/
Crawling sub.. https://golang.org/pkg/fmt/
Crawling sub.. https://golang.org/pkg/os/
Skipping ...  https://golang.org/cmd/
Skipping ...  https://golang.org/
Fetching: https://golang.org/pkg/os/ ...
Fetching: https://golang.org/pkg/fmt/ ...
Found:  https://golang.org/pkg/fmt/ Package fmt
Crawling sub.. https://golang.org/
Crawling sub.. https://golang.org/pkg/
Skipping ...  https://golang.org/pkg/
Master Status +1 :  https://golang.org/pkg/fmt/
Found:  https://golang.org/pkg/os/ Package os
Crawling sub.. https://golang.org/
Crawling sub.. https://golang.org/pkg/
Skipping ...  https://golang.org/pkg/
Master Status +1 :  https://golang.org/pkg/os/
Skipping ...  https://golang.org/
Skipping ...  https://golang.org/
*/
