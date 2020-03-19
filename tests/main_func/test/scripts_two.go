package main

import (
	"fmt"
	"strconv"
	"strings"
)

func StringSliceScript(s []string, sp []*string, ps *[]string, psp *[]*string) {
	sOut := strings.Join(s, " ")
	spOutSlice := make([]string, 0, len(sp))
	for _, el := range sp {
		if el == nil {
			spOutSlice = append(spOutSlice, "nil")
			continue
		}
		spOutSlice = append(spOutSlice, *el)
	}
	spOut := strings.Join(spOutSlice, " ")
	var psOut string
	if ps == nil {
		psOut = "nil slice"
	} else {
		psOutSlice := make([]string, 0)
		for _, el := range *ps {
			psOutSlice = append(psOutSlice, el)
		}
		psOut = strings.Join(psOutSlice, " ")
	}
	var pspOut string
	if psp == nil {
		pspOut = "nil slice"
	} else {
		pspOutSlice := make([]string, 0)
		for _, el := range *psp {
			if el == nil {
				pspOutSlice = append(pspOutSlice, "nil")
			} else {
				pspOutSlice = append(pspOutSlice, *el)
			}
		}
		pspOut = strings.Join(pspOutSlice, " ")
	}
	fmt.Printf("s: %s\nsp: %s\nps: %s\npsp: %s", sOut, spOut, psOut, pspOut)
}

func IntSliceScript(n []int64, np []*int64, pn *[]int64, pnp *[]*int64) {
	nStr := make([]string, len(n))
	for i, el := range n {
		nStr[i] = strconv.FormatInt(el, 10)
	}
	nOut := strings.Join(nStr, " ")
	npOutSlice := make([]string, 0, len(np))
	for _, el := range np {
		if el == nil {
			npOutSlice = append(npOutSlice, "nil")
			continue
		}
		npOutSlice = append(npOutSlice, strconv.FormatInt(*el, 10))
	}
	npOut := strings.Join(npOutSlice, " ")
	var pnOut string
	if pn == nil {
		pnOut = "nil slice"
	} else {
		pnOutSlice := make([]string, 0)
		for _, el := range *pn {
			pnOutSlice = append(pnOutSlice, strconv.FormatInt(el, 10))
		}
		pnOut = strings.Join(pnOutSlice, " ")
	}
	var pnpOut string
	if pnp == nil {
		pnpOut = "nil slice"
	} else {
		pnpOutSlice := make([]string, 0)
		for _, el := range *pnp {
			if el == nil {
				pnpOutSlice = append(pnpOutSlice, "nil")
			} else {
				pnpOutSlice = append(pnpOutSlice, strconv.FormatInt(*el, 10))
			}
		}
		pnpOut = strings.Join(pnpOutSlice, " ")
	}
	fmt.Printf("n: %s\nnp: %s\npn: %s\npnp: %s", nOut, npOut, pnOut, pnpOut)
}

func BoolSliceScript(b []bool, bp []*bool, pb *[]bool, pbp *[]*bool) {
	bStr := make([]string, len(b))
	for i, el := range b {
		bStr[i] = fmt.Sprintf("%t", el)
	}
	bOut := strings.Join(bStr, " ")
	bpOutSlice := make([]string, 0, len(bp))
	for _, el := range bp {
		if el == nil {
			bpOutSlice = append(bpOutSlice, "nil")
			continue
		}
		bpOutSlice = append(bpOutSlice, fmt.Sprintf("%t", *el))
	}
	bpOut := strings.Join(bpOutSlice, " ")
	var pbOut string
	if pb == nil {
		pbOut = "nil slice"
	} else {
		pbOutSlice := make([]string, 0)
		for _, el := range *pb {
			pbOutSlice = append(pbOutSlice, fmt.Sprintf("%t", el))
		}
		pbOut = strings.Join(pbOutSlice, " ")
	}
	var pbpOut string
	if pbp == nil {
		pbpOut = "nil slice"
	} else {
		pbpOutSlice := make([]string, 0)
		for _, el := range *pbp {
			if el == nil {
				pbpOutSlice = append(pbpOutSlice, "nil")
			} else {
				pbpOutSlice = append(pbpOutSlice, fmt.Sprintf("%t", *el))
			}
		}
		pbpOut = strings.Join(pbpOutSlice, " ")
	}
	fmt.Printf("b: %s\nbp: %s\npb: %s\npbp: %s", bOut, bpOut, pbOut, pbpOut)
}

func StringArrScript(s [3]string, sp [3]*string, ps *[3]string, psp *[3]*string) {
	replaceEmptyFn := func(s string) string {
		if s == "  " {
			return ""
		}
		return s
	}
	sOutSlice := s[:]
	sOut := strings.Join(sOutSlice, " ")
	sOut = replaceEmptyFn(sOut)
	spOutSlice := make([]string, 0, len(sp))
	for _, el := range sp {
		if el == nil {
			spOutSlice = append(spOutSlice, "nil")
			continue
		}
		spOutSlice = append(spOutSlice, *el)
	}
	spOut := strings.Join(spOutSlice, " ")
	spOut = replaceEmptyFn(spOut)
	var psOut string
	if ps == nil {
		psOut = "nil slice"
	} else {
		psOutSlice := make([]string, 0)
		for _, el := range *ps {
			psOutSlice = append(psOutSlice, el)
		}
		psOut = strings.Join(psOutSlice, " ")
		psOut = replaceEmptyFn(psOut)
	}
	var pspOut string
	if psp == nil {
		pspOut = "nil slice"
	} else {
		pspOutSlice := make([]string, 0)
		for _, el := range *psp {
			if el == nil {
				pspOutSlice = append(pspOutSlice, "nil")
			} else {
				pspOutSlice = append(pspOutSlice, *el)
			}
		}
		pspOut = strings.Join(pspOutSlice, " ")
		pspOut = replaceEmptyFn(pspOut)
	}
	fmt.Printf("s: %s\nsp: %s\nps: %s\npsp: %s", sOut, spOut, psOut, pspOut)
}

func StringArrLengthOneScript(s [1]string) {
	replaceEmptyFn := func(s string) string {
		if s == "  " {
			return ""
		}
		return s
	}
	sOutSlice := s[:]
	sOut := strings.Join(sOutSlice, " ")
	sOut = replaceEmptyFn(sOut)
	fmt.Printf("s: %s", sOut)
}
