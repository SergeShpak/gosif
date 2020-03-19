package main

import "fmt"

type SomeStruct struct {
	payload int
}

func StructScript(s SomeStruct) {
	fmt.Printf("payload: %d", s.payload)
}

func unexportedScript(n int) {
	fmt.Printf("n: %d", n)
}
