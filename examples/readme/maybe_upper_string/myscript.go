package main

import (
	"fmt"
	"strings"
)

func PrintStringMaybeUpper(parts []string, upper bool) {
	str := strings.Join(parts, " ")
	if upper {
		str = strings.ToUpper(str)
	}
	fmt.Println(str)
}
