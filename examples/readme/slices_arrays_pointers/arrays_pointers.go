package main

import (
	"fmt"
)

func MyArrPointerFunc(pa *[3]int, ap [3]*int, pap *[3]*int) {
	if pa == nil {
		fmt.Println("pa: nil")
	} else {
		fmt.Println("pa: ", *pa)
	}
	apElems := make([]string, len(ap))
	for i, el := range ap {
		if el == nil {
			apElems[i] = "nil"
		} else {
			apElems[i] = fmt.Sprintf("%d", *el)
		}
	}
	fmt.Println("ap: ", apElems)
	if pap == nil {
		fmt.Println("pap: nil")
	} else {
		papElems := make([]string, len(*pap))
		for i, el := range *pap {
			if el == nil {
				papElems[i] = "nil"
			} else {
				papElems[i] = fmt.Sprintf("%d", *el)
			}
		}
		fmt.Println("pap: ", papElems)
	}
}
