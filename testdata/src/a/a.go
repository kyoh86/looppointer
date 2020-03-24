package main

func main() {
	var list []*int
	var ref *int

	println("loop")
	for i, p := range []int{10, 11, 12, 13} {
		printp(&p)              // not a diagnostic
		list = append(list, &p) // want "collecting pointers to a loop variable p"
		if i%2 == 0 {
			ref = &p // want "capturing a pointer to a loop variable p"
		}
	}

	println("expand collection")
	for _, p := range list {
		printp(p) // expecting "10, 11, 12, 13" but "13, 13, 13, 13"
	}
	println("expand captured value")
	printp(ref) // expecting "12" but "13"
}

func printp(p *int) {
	println(*p)
}
