package main

import "fmt"

func StringType(s string) {
	fmt.Println(s)
}

func ByteType(b byte) {
	fmt.Println(b)
}

func RuneType(r rune) {
	fmt.Println(string(r))
}

func BoolType(bt bool, bf bool) {
	fmt.Println(bt, bf)
}

func SignedIntType(n int, n8 int8, n16 int16, n32 int32, n64 int64) {
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n", n, n8, n16, n32, n64)
}

func UnsignedIntType(n uint, n8 uint8, n16 uint16, n32 uint32, n64 uint64) {
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n", n, n8, n16, n32, n64)
}

func FloatType(f32 float32, f64 float64) {
	fmt.Printf("f32: %.3f, f64: %.3f\n", f32, f64)
}

func ComplexType(c64 complex64, c128 complex128) {
	fmt.Printf("c64: (%.3f, %.3fi), c128: (%.3f, %.3fi)\n", real(c64), imag(c64), real(c128), imag(c128))
}

func ErrorType(e error) {
	fmt.Println(e)
}
