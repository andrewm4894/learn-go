package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type NetdataResponse struct {
	Labels []string    `json:"labels"`
	Data   [][]float64 `json:"data"`
}

func getNetdataResponse(url string) NetdataResponse {
	var dataMap NetdataResponse
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &dataMap)
	return dataMap
}

func main() {

	urls := []string{
		"https://london.my-netdata.io/api/v1/data?chart=system.cpu&after=-5",
		"https://london.my-netdata.io/api/v1/data?chart=system.net&after=-5",
		"https://london.my-netdata.io/api/v1/data?chart=system.load&after=-5",
		"https://london.my-netdata.io/api/v1/data?chart=system.io&after=-5",
	}

	for _, url := range urls {

		data := getNetdataResponse(url)
		fmt.Println(data.Labels)
		fmt.Println(data.Data)

	}

}
