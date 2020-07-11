package main

func main() {
	var intSlice []*int

	println("loop expecting 10, 11, 12, 13")
	for _, p := range []int{10, 11, 12, 13} {
		intSlice = append(intSlice, &p) // nolint:looppointer
	}

	println(`slice expecting "10, 11, 12, 13" but "13, 13, 13, 13"`)
	for _, p := range intSlice {
		printp(p)
	}
}

func printp(p *int) {
	println(*p)
}
