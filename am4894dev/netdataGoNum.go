package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/sjwhitworth/golearn/trees"
	"gonum.org/v1/gonum/mat"
)

// Create a wait group
var wg sync.WaitGroup

// LagsN defines number of lags to make
var LagsN = 2

// Urls define a list of api calls we want data from
var Urls = [1]string{
	//"https://london.my-netdata.io/api/v1/data?chart=system.cpu&format=json&after=-4",
	"https://london.my-netdata.io/api/v1/data?chart=system.net&format=json&after=-20",
	//"https://london.my-netdata.io/api/v1/data?chart=system.load&format=json&after=-4",
	//"https://london.my-netdata.io/api/v1/data?chart=system.io&format=json&after=-3",
}

type netdataResponse struct {
	Labels []string    `json:"labels"`
	Data   [][]float64 `json:"data"`
}

// Get a gonum matrix from the netdata api with specified nLags
func getX(url string, nLags int, c chan mat.Matrix) {

	// Create data to store response
	var data netdataResponse

	// Need to make sure we tell wait group we done
	defer wg.Done()

	// Get response
	resp, _ := http.Get(url)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Unmarshal into dataMap
	_ = json.Unmarshal([]byte(bodyBytes), &data)

	// Flatten data into one slice, ignoring the first column which is "time", and adding nLags
	nData := len(data.Data)
	nDims := len(data.Labels) - 1
	nCols := nDims + (nLags * nDims)
	nRows := nData - nLags

	// Make flat slice to put data into
	dataFlat := make([]float64, nCols*nRows)

	// Loop over and add lags to flat data
	i := 0
	for t := range data.Data {
		fmt.Println(data.Data[t])
		if t >= nLags {
			for dim := range data.Data[t] {
				// Ignore time which is the first dim in the response
				if dim > 0 {
					// Add each lag
					for l := 0; l <= nLags; l++ {
						dataFlat[i] = data.Data[t-l][dim]
						i++
					}
				}
			}
		}
	}

	// Create gonum dense matrix from dataFlat
	x := mat.NewDense(nRows, nCols, dataFlat)

	// Send to channel
	c <- x

}

func main() {

	// Create a channel the size of number of api calls we need to make
	dataChannel := make(chan mat.Matrix, len(Urls))

	// Kick off a go routine for each url
	for _, url := range Urls {
		wg.Add(1)
		go getX(url, LagsN, dataChannel)
	}

	// Handle synchronization of channel
	wg.Wait()
	close(dataChannel)

	// Pull each response from channel
	for x := range dataChannel {

		// Print matrix x
		fmt.Printf("X:\n %v\n\n", mat.Formatted(x, mat.Prefix(" "), mat.Excerpt(10)))

		fmt.Println(x.Dims)
		//instances := base.InstancesFromMat64(1, 1, x)

		forest := trees.NewIsolationForest(10, 10, 10)
		forest.Fit(x)
		preds := forest.Predict(x)

		// Let's find the average and minimum Anomaly Score for normal data
		var avgScore float64
		var min float64
		min = 1

		for i := 0; i < 1000; i++ {
			temp := preds[i]
			avgScore += temp
			if temp < min {
				min = temp
			}
		}
		fmt.Println(avgScore / 1000)
		fmt.Println(min)

		fmt.Println("Anomaly Scores are ")
		for i := range preds {
			fmt.Println(preds[i])
		}

	}

}
