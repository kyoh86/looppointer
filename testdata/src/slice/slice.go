package main

import "fmt"

var a = [][3]int{
	{1, 2, 3},
	{4, 5, 6},
	{7, 8, 9},
}

func main() {
	var s [][]int
	for _, e := range a {
		s = append(s, e[:]) // want "taking a ref for the slice from loop variable e"
	}
	fmt.Println(s)
}
