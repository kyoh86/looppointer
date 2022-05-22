package main

import "time"

func main() {
	var intPointerSlice []*int
	var unrelatedIntSlice []int

	for _, p := range []int{10, 11, 12, 13} {
		go func() {
			// Ensure not all loop variables in a func literal are flagged
			for _, ui := range []int{1, 2, 3} {
				unrelatedIntSlice = append(unrelatedIntSlice, ui)
			}

			// Double trouble! Taking a loop pointer happens to be reported first.
			intPointerSlice = append(intPointerSlice, &p) // want "taking a pointer for the loop variable p"
		}()
	}

	var intSlice []int // The func literal bug does not require a pointer

	for _, p := range []int{10, 11, 12, 13} {
		go func() {
			intSlice = append(intSlice, p) // want "using loop variable in function literal p"
		}()
	}

	for _, p := range []int{10, 11, 12, 13} {
		go func() {
			go func() {
				for _, ui := range []int{1, 2, 3} {
					unrelatedIntSlice = append(unrelatedIntSlice, ui)
				}

				intSlice = append(intSlice, p) // want "using loop variable in function literal p"
			}()
		}()
	}

	// Ensure everything within a literal isn't ignored
	go func() {
		for _, p := range []int{10, 11, 12, 13} {
			go func() {
				for _, ui := range []int{1, 2, 3} {
					unrelatedIntSlice = append(unrelatedIntSlice, ui)
				}

				intSlice = append(intSlice, p) // want "using loop variable in function literal p"
			}()
		}
	}()

	// This is materially no different from the other basic tests, but shows that it's not just "go" func that's an
	// issue.
	var funcSlice []func()
	for _, p := range []int{10, 11, 12, 13} {
		funcSlice = append(funcSlice, func() {
			intSlice = append(intSlice, p) // want "using loop variable in function literal p"
		})
	}

	// This actually does NOT exhibit the bug and is a false positive. Creating a func literal and executing it
	// sync should be relatively uncommon?
	for _, p := range []int{10, 11, 12, 13} {
		func() {
			intSlice = append(intSlice, p) // want "using loop variable in function literal p"
		}()
	}

	// sync.WaitGroup's lazy cousin
	time.Sleep(time.Millisecond)
}
