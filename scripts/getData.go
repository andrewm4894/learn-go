package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type urlResponses struct {
	url    string
	status bool
	data   string
}

func main() {
	urls := []string{
		"https://london.my-netdata.io/api/v1/data?chart=system.cpu&after=-10",
		"https://london.my-netdata.io/api/v1/data?chart=system.net&after=-10",
	}

	c := make(chan urlResponses)
	for _, url := range urls {
		go getData(url, c)

	}
	result := make([]urlResponses, len(urls))
	for i, _ := range result {
		result[i] = <-c
		if result[i].status {
			fmt.Println(result[i].data)
			//dataMap := map[string]string{}
			//err := json.Unmarshal(result[i].data, &dataMap)
			//if err != nil {
			//	log.Fatal(err.Error())
			//}
			//fmt.Println(dataMap)
		}
	}

}

func getData(url string, c chan urlResponses) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		c <- urlResponses{url, true, bodyString}
		//c <- urlResponses{url, true, bodyBytes}
	}
}
