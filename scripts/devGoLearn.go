package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/sjwhitworth/golearn/base"
)

// Create a wait group
var wg sync.WaitGroup

// Get api response (expects format=csv) and make a dataframe from it
func getDf(url string, c chan dataframe.DataFrame) {

	// Need to make sure we tell wait group we done
	defer wg.Done()

	// Pull chart name from the url
	re := regexp.MustCompile("chart=(.*?)&")
	match := re.FindStringSubmatch(url)
	chart := match[1]
	resp, _ := http.Get(url)

	// Get body as string for ReadCSV
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	df := dataframe.ReadCSV(strings.NewReader(bodyString))

	// Add chart suffix to each col name
	// (ignore first col which should be "time" and used for joins later)
	colnames := df.Names()
	for i, colname := range colnames {
		if i != 0 {
			df = df.Rename(chart+"|"+colname, colname)
		}
	}

	// send df to channel
	c <- df

}

func main() {

	// Define a list of api calls we want data from
	// In this example we have an api call for each chart data we want in our df
	urls := []string{
		"https://london.my-netdata.io/api/v1/data?chart=system.cpu&format=csv&after=-100",
		"https://london.my-netdata.io/api/v1/data?chart=system.io&format=csv&after=-100",
	}

	// Create a channel of dataframes the size of number of api calls we need to make
	dfChannel := make(chan dataframe.DataFrame, len(urls))

	// Create empty df we will outer join into from the df channel later
	df := dataframe.ReadJSON(strings.NewReader(`[{"time":"1900-01-01 00:00:01"}]`))

	// Kick off a go routine for each url
	for _, url := range urls {
		wg.Add(1)
		go getDf(url, dfChannel)
	}

	// Handle synchronization of channel
	wg.Wait()
	close(dfChannel)

	// Pull each df from the channel and outer join onto our original empty df
	for dfTmp := range dfChannel {
		df = df.OuterJoin(dfTmp, "time")
	}

	// Sort based on time
	df = df.Arrange(dataframe.Sort("time"))

	// Drop dummy time row
	df = df.Filter(
		dataframe.F{
			Colname:    "time",
			Comparator: series.Neq,
			Comparando: "1900-01-01 00:00:01",
		},
	)

	// Print df
	fmt.Println(df, 10, 5)

	dataInstances, err := base.ParseCSVToInstances(df, true)
	if err != nil {
		panic(err)
	}

	// Describe df
	//fmt.Println(df.Describe())

}
