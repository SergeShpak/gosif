package main

import (
	"fmt"
	"strings"
)

func SimpleScript(arg string) {
	fmt.Print(arg)
}

func IntScript(n int, np *int, n8 int8, n8p *int8, n16 int16, n16p *int16, n32 int32, n32p *int32, n64 int64, n64p *int64) {
	pointers := make([]string, 5)
	if np == nil {
		pointers[0] = "np: nil"
	} else {
		pointers[0] = fmt.Sprintf("np: %d", *np)
	}
	if n8p == nil {
		pointers[1] = "n8p: nil"
	} else {
		pointers[1] = fmt.Sprintf("n8p: %d", *n8p)
	}
	if n16p == nil {
		pointers[2] = "n16p: nil"
	} else {
		pointers[2] = fmt.Sprintf("n16p: %d", *n16p)
	}
	if n32p == nil {
		pointers[3] = "n32p: nil"
	} else {
		pointers[3] = fmt.Sprintf("n32p: %d", *n32p)
	}
	if n64p == nil {
		pointers[4] = "n64p: nil"
	} else {
		pointers[4] = fmt.Sprintf("n64p: %d", *n64p)
	}
	pointersStr := strings.Join(pointers, ", ")
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n%s", n, n8, n16, n32, n64, pointersStr)
}

func UintScript(n uint, np *uint, n8 uint8, n8p *uint8, n16 uint16, n16p *uint16, n32 uint32, n32p *uint32, n64 uint64, n64p *uint64) {
	pointers := make([]string, 5)
	if np == nil {
		pointers[0] = "np: nil"
	} else {
		pointers[0] = fmt.Sprintf("np: %d", *np)
	}
	if n8p == nil {
		pointers[1] = "n8p: nil"
	} else {
		pointers[1] = fmt.Sprintf("n8p: %d", *n8p)
	}
	if n16p == nil {
		pointers[2] = "n16p: nil"
	} else {
		pointers[2] = fmt.Sprintf("n16p: %d", *n16p)
	}
	if n32p == nil {
		pointers[3] = "n32p: nil"
	} else {
		pointers[3] = fmt.Sprintf("n32p: %d", *n32p)
	}
	if n64p == nil {
		pointers[4] = "n64p: nil"
	} else {
		pointers[4] = fmt.Sprintf("n64p: %d", *n64p)
	}
	pointersStr := strings.Join(pointers, ", ")
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n%s", n, n8, n16, n32, n64, pointersStr)
}

func FloatScript(n32 float32, n32p *float32, n64 float64, n64p *float64) {
	pointers := make([]string, 2)
	if n32p == nil {
		pointers[0] = "n32p: nil"
	} else {
		pointers[0] = fmt.Sprintf("n32p: %.3f", *n32p)
	}
	if n64p == nil {
		pointers[1] = "n64p: nil"
	} else {
		pointers[1] = fmt.Sprintf("n64p: %.3f", *n64p)
	}
	pointersStr := strings.Join(pointers, ", ")
	fmt.Printf("n32: %.3f, n64: %.3f\n%s", n32, n64, pointersStr)
}

func StringScript(s string, sp *string) {
	var pointerStr string
	if sp == nil {
		pointerStr = "sp: nil"
	} else {
		pointerStr = fmt.Sprintf("sp: %s", *sp)
	}
	fmt.Printf("s: %s\n%s", s, pointerStr)
}

func ByteScript(b byte, bp *byte) {
	var pointerStr string
	if bp == nil {
		pointerStr = "bp: nil"
	} else {
		pointerStr = fmt.Sprintf("bp: %d", *bp)
	}
	fmt.Printf("b: %d\n%s", b, pointerStr)
}

func RuneScript(r rune, rp *rune) {
	var pointerStr string
	if rp == nil {
		pointerStr = "rp: nil"
	} else {
		pointerStr = fmt.Sprintf("rp: %s", string(*rp))
	}
	fmt.Printf("r: %s\n%s", string(r), pointerStr)
}

func BoolScript(b bool, bp *bool) {
	var boolStr string
	if bp == nil {
		boolStr = "bp: nil"
	} else {
		boolStr = fmt.Sprintf("bp: %t", *bp)
	}
	fmt.Printf("b: %t\n%s", b, boolStr)
}

func ComplexScript(c64 complex64, c64p *complex64, c128 complex128, c128p *complex128) {
	complexToStrFn := func(c complex128) string {
		realPart := real(c)
		imgPart := imag(c)
		return fmt.Sprintf("(%.3f,%.3fi)", realPart, imgPart)
	}
	pointers := make([]string, 2)
	if c64p == nil {
		pointers[0] = "c64p: nil"
	} else {
		pointers[0] = fmt.Sprintf("c64p: %s", complexToStrFn(complex128(*c64p)))
	}
	if c128p == nil {
		pointers[1] = "c128p: nil"
	} else {
		pointers[1] = fmt.Sprintf("c128p: %s", complexToStrFn(*c128p))
	}
	pointersStr := strings.Join(pointers, ", ")
	fmt.Printf("c64: %s, c128: %s\n%s", complexToStrFn(complex128(c64)), complexToStrFn(c128), pointersStr)
}

func ErrorScript(e error, ep *error) {
	pointerStr := "nil"
	if ep != nil {
		pointerStr = fmt.Sprintf("ep: %s", (*ep).Error())
	}
	fmt.Printf("e: %s\n%s", e.Error(), pointerStr)
}

func NoArgsScript() {
	fmt.Printf("this script does not expect any args")
}
