package main

import "fmt"

func MySlicePointerFunc(ps *[]int, sp []*int, psp *[]*int) {
	if ps == nil {
		fmt.Println("ps: nil")
	} else {
		fmt.Println("ps: ", *ps)
	}
	spElems := make([]int, len(sp))
	for i, el := range sp {
		spElems[i] = *el
	}
	fmt.Println("sp: ", spElems)
	if psp == nil {
		fmt.Println("psp: nil")
	} else {
		pspElems := make([]int, len(*psp))
		for i, el := range *psp {
			pspElems[i] = *el
		}
		fmt.Println("psp: ", pspElems)
	}
}
