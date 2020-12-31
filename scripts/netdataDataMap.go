package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"gonum.org/v1/gonum/mat"
)

// Create a wait group
var wg sync.WaitGroup

type netdataResponse struct {
	Labels []string    `json:"labels"`
	Data   [][]float64 `json:"data"`
}

// Get data from api
func getData(url string, c chan netdataResponse) {

	// Create data to store response
	var data netdataResponse

	// Need to make sure we tell wait group we done
	defer wg.Done()

	// Get response
	resp, _ := http.Get(url)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Unmarshal into dataMap
	_ = json.Unmarshal([]byte(bodyBytes), &data)

	// Send to channel
	c <- data

}

func main() {

	// Define a list of api calls we want data from
	urls := []string{
		//"https://london.my-netdata.io/api/v1/data?chart=system.cpu&format=json&after=-4",
		"https://london.my-netdata.io/api/v1/data?chart=system.net&format=json&after=-20",
		//"https://london.my-netdata.io/api/v1/data?chart=system.load&format=json&after=-4",
		//"https://london.my-netdata.io/api/v1/data?chart=system.io&format=json&after=-3",
	}

	// Create a channel the size of number of api calls we need to make
	dataChannel := make(chan netdataResponse, len(urls))

	// Kick off a go routine for each url
	for _, url := range urls {
		wg.Add(1)
		go getData(url, dataChannel)
	}

	// Handle synchronization of channel
	wg.Wait()
	close(dataChannel)

	// Pull each response from channel
	for data := range dataChannel {

		// Flatten data into one slice, ignoring the first column which is "time", and adding lags_n
		nLags := 3
		nData := len(data.Data)
		nDims := len(data.Labels) - 1
		nCols := nDims + (nLags * nDims)
		nRows := nData - nLags
		fmt.Println(nCols)
		fmt.Println(nRows)
		dataFlat := make([]float64, nCols*nRows)
		i := 0
		for t := range data.Data {
			fmt.Println(data.Data[t])
			if t >= nLags {
				for dim := range data.Data[t] {
					if dim > 0 {
						for l := 0; l <= nLags; l++ {
							dataFlat[i] = data.Data[t-l][dim]
							i++
						}
					}
				}
			}
		}

		fmt.Printf("%v\n", dataFlat)

		// Create gonum dense matrix from dataFlat
		X := mat.NewDense(nRows, nCols, dataFlat)

		// Print matrix X
		fmt.Printf("X:\n %v\n\n", mat.Formatted(X, mat.Prefix(" "), mat.Excerpt(10)))

	}

}
