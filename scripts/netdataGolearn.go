package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/trees"
	"gonum.org/v1/gonum/mat"
)

// Create a wait group
var wg sync.WaitGroup

type netdataResponse struct {
	Labels []string    `json:"labels"`
	Data   [][]float64 `json:"data"`
}

// Get a instances from the netdata api
func getInstances(conf map[string]interface{}, c chan base.FixedDataGrid) {

	url := "https://" + conf["host"].(string) + "/api/v1/data?chart=" + conf["chart"].(string) + "&format=json&after=" + conf["trainAfter"].(string) + "&before=" + conf["trainBefore"].(string)
	lags := conf["lags"].(int)
	fmt.Println(url)

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
	nCols := nDims + (lags * nDims)
	nRows := nData - lags

	// Make flat slice to put data into
	dataFlat := make([]float64, nCols*nRows)

	// Loop over and add lags to flat data
	i := 0
	for t := range data.Data {
		fmt.Println(data.Data[t])
		if t >= lags {
			for dim := range data.Data[t] {
				// Ignore time which is the first dim in the response
				if dim > 0 {
					// Add each lag
					for l := 0; l <= lags; l++ {
						dataFlat[i] = data.Data[t-l][dim]
						i++
					}
				}
			}
		}
	}

	// Create gonum dense matrix from dataFlat
	x := mat.NewDense(nRows, nCols, dataFlat)

	// Create vector of zeros to use as a dummy class attribute for golearn
	var xFinal mat.Dense
	zeros := make([]float64, nRows)
	z := mat.NewVecDense(nRows, zeros)
	xFinal.Augment(z, x)

	// Create instances
	nrow, ncol := xFinal.Dims()
	instances := base.NewDenseCopy(base.InstancesFromMat64(nrow, ncol, &xFinal))

	// Set a class attribute
	attrArray := instances.AllAttributes()
	instances.AddClassAttribute(attrArray[0])

	// Send to channel
	c <- instances

}

func main() {

	// define config
	var host = "london.my-netdata.io"
	var trainAfter = "-10"
	var trainBefore = "0"
	var lags = 2
	config := map[string]map[string]interface{}{
		"1": {"host": host, "chart": "system.net", "trainAfter": trainAfter, "trainBefore": trainBefore, "lags": lags},
	}

	// Create a channel the size of number of api calls we need to make
	trainDataChannel := make(chan base.FixedDataGrid, len(config))

	// Kick off a go routine for each url
	for _, conf := range config {
		wg.Add(1)
		go getInstances(conf, trainDataChannel)
	}

	// Handle synchronization of channel
	wg.Wait()
	close(trainDataChannel)

	// Pull each response from channel
	for instances := range trainDataChannel {

		fmt.Println(instances)

		// Create forest
		forest := trees.NewIsolationForest(10, 10, 100)

		// Fit forest
		forest.Fit(instances)

		// Predict on instances to get scores
		preds := forest.Predict(instances)

		// Let's find the average and minimum Anomaly Score for normal data
		var avgScore float64
		var min float64
		min = 1
		for i := 0; i < len(preds); i++ {
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
