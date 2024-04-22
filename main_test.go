package main

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	a := "A"
	b := "A"
	ax := &a
	bx := &b
	fmt.Println(a == b)
	fmt.Println(ax == bx)
}
