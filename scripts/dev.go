package main

import (
	"fmt"
)

func main() {
	nSmooth := 2
	fmt.Println(nSmooth)
	x := [][]float64{1., 2., 3., 4., 5., 6., 7., 8., 9., 10.}
	fmt.Println(x)
	//for i := range x {
	//	if i > nSmooth {
	//		fmt.Println(stat.Mean(x[i-nSmooth:i], nil))
	//	}
	//}
	//mean := stat.Mean(x, nil)
	//fmt.Printf("The mean of x is %.4f\n", mean)
}
