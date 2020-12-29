package main

import (
	"context"
	"fmt"
	"strings"

	imports "github.com/rocketlaunchr/dataframe-go"
)

func main() {

	csvStr := `colA,colB
	1,"First"
	2,"Second"
	3,"Third"
	4,"Fourth"`

	ctx := context.Background()

	df, err := imports.LoadFromCSV(ctx, strings.NewReader(csvStr), imports.CSVLoadOptions{
		DictateDataType: map[string]interface{}{
			"colA": int64(0),
			"colB": "",
		},
	})

	fmt.Println(err)
	fmt.Println(df)

}
