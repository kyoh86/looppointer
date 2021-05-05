package main

func main() {
	var ref *int

	for _, v := range []int{10, 11, 12, 13} {
		if v == 11 {
			ref = &v // want "taking a pointer for the loop variable v"
		}
	}
	printp(ref)
}

func printp(p *int) {
	println(*p)
}
