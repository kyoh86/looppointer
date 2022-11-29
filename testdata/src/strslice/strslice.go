package main

import "fmt"

var a = []string{
	"123",
	"456",
	"789",
}

type strlike string

var b = []string{
	"abc",
	"def",
	"ghi",
}

func main() {
	var s []string
	for _, e := range a {
		s = append(s, e[:1]) // no problem
	}
	for _, e := range b {
		s = append(s, e[:1]) // no problem
	}

	fmt.Printf("%q", s)
}
