# looppointer

An analyzer that finds pointers for loop variables.

[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/looppointer)](https://goreportcard.com/report/github.com/kyoh86/looppointer)
[![Coverage Status](https://img.shields.io/codecov/c/github/kyoh86/looppointer.svg)](https://codecov.io/gh/kyoh86/looppointer)

## What's this?

Sample problem code from: https://github.com/kyoh86/looppointer/blob/master/testdata/simple/simple.go

```go
package main

func main() {
	var intSlice []*int

	println("loop expecting 10, 11, 12, 13")
	for _, p := range []int{10, 11, 12, 13} {
		intSlice = append(intSlice, &p) // want "taking a pointer for the loop variable p"
	}

	println(`slice expecting "10, 11, 12, 13" but "13, 13, 13, 13"`)
	for _, p := range intSlice {
		printp(p)
	}
}

func printp(p *int) {
	println(*p)
}
```

In Go, the `p` variable in the above loops is actually a single variable.
So in many case (like the above), using it makes for us annoying bugs.

You can find them with `looppointer`, and fix it.

```go
package main

func main() {
	var intSlice []*int

	println("loop expecting 10, 11, 12, 13")
	for i, p := range []int{10, 11, 12, 13} {
    p := p                          // FIX variable into the inner variable
		intSlice = append(intSlice, &p) 
	}

	println(`slice expecting "10, 11, 12, 13"`)
	for _, p := range intSlice {
		printp(p)
	}
}

func printp(p *int) {
	println(*p)
}
```

ref: https://github.com/kyoh86/looppointer/blob/master/testdata/fixed/fixed.go

## Sensing policy

I want to make looppointer as nervous as possible.
So many false-positves will be reported.

e.g.

```go
func TestSample(t *testing.T) {
  for _, p := []int{10, 11, 12, 13} {
    t.Run(func(t *testing.T) {
      s = &p // t.Run always called instantly, so it will not be bug.
      ...
    })
  }
}
```

If you want to ignore false-positives (with some lints ignored),
you should use [exportloopref](https://github.com/kyoh86/exportloopref).

## Install

go:

```console
$ go get github.com/kyoh86/looppointer/cmd/looppointer
```

[homebrew](https://brew.sh/):

```console
$ brew install kyoh86/tap/looppointer
```

[gordon](https://github.com/kyoh86/gordon):

```console
$ gordon install kyoh86/looppointer
```

## Usage

```
looppointer [-flag] [package]
```

### Flags

| Flag | Description |
| --- | --- |
| -V                 | print version and exit |
| -all               | no effect (deprecated) |
| -c int             | display offending line with this many lines of context (default -1) |
| -cpuprofile string | write CPU profile to this file |
| -debug string      | debug flags, any subset of "fpstv" |
| -fix               | apply all suggested fixes |
| -flags             | print analyzer flags in JSON |
| -json              | emit JSON output |
| -memprofile string | write memory profile to this file |
| -source            | no effect (deprecated) |
| -tags string       | no effect (deprecated) |
| -trace string      | write trace log to this file |
| -v                 | no effect (deprecated) |

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This is distributed under the [MIT License](http://www.opensource.org/licenses/MIT).
