package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	urls := []string{
		"https://london.my-netdata.io/api/v1/data?chart=system.cpu&after=-2",
		"https://london.my-netdata.io/api/v1/data?chart=system.net&after=-2",
		"https://london.my-netdata.io/api/v1/data?chart=system.load&after=-2",
		"https://london.my-netdata.io/api/v1/data?chart=system.io&after=-2",
	}

	jobs := make(chan string, len(urls))
	results := make(chan string, len(urls))

	go worker(jobs, results)
	go worker(jobs, results)

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	for j := 0; j < len(urls); j++ {
		fmt.Println(<-results)
	}
}

func worker(jobs <-chan string, results chan<- string) {
	for url := range jobs {
		results <- getData(url)
	}
}

func getData(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString
}
