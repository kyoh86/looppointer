package main

type A struct {
	b B
}
type B struct {
	c C
}
type C struct {
	d int
}

func NewA(value int) A {
	return A{B{C{value}}}
}

func main() {
	var intSlice []*int

	println("loop expecting 1, 2, 3, 4")
	for _, p := range []A{NewA(1), NewA(2), NewA(3), NewA(4)} {
		intSlice = append(intSlice, &p.b.c.d) // want "taking a pointer for the loop variable p"
	}

	println(`slice expecting "1, 2, 3, 4" but "4, 4, 4, 4"`)
	for _, p := range intSlice {
		printp(p)
	}
}

func printp(p *int) {
	println(*p)
}
