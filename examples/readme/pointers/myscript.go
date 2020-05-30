package main

import "fmt"

func MyPointerFunc(num *int) {
	if num == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(*num)
	}
}

func MyManyPointersFunc(num ***int) {
	if num == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(***num)
	}
}

func BoolPointerFunc(b *bool) {
	if b == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(*b)
	}
}
