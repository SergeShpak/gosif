package main

import "fmt"

func MySliceFunc(nums []int) { fmt.Println(nums) }
func MyArrFunc(nums [3]int)  { fmt.Println(nums) }

// multi-dimensional slices and arrays are currently not supported
func MyMultiSliceFunc(nums [][][]int) { fmt.Println(nums) }
func MyMultiArrFunc(nums [2][3]int)   { fmt.Println(nums) }
