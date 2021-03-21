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

func ShortNamesScript(alpha string, beta string, gamma string, betaTwo string) {
	fmt.Printf("alpha: %s, beta: %s, gamma: %s, betaTwo: %s", alpha, beta, gamma, betaTwo)
}
