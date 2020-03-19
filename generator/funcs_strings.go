package generator

import (
	"fmt"
	"strings"
)

type predefinedFunc struct {
	name string
	body string
}

var funcStringParseArgAsComplex predefinedFunc = predefinedFunc{
	name: "funcStringParseArgAsComplex",
	body: `type gosif_ComplexArgStruct struct {
		R string
		I string
	}
	
	func gosif_ParseArgAsComplex(arg string) (*gosif_ComplexArgStruct, error) {
		if len(arg) == 0 {
			return nil, fmt.Errorf("an empty string cannot be parsed as a complex number")
		}
		generateFormatErrMsgFn := func(reason string) error {
			return fmt.Errorf("\"%s\" cannot be parsed as a complex number: %s", arg, reason)
		}
		normalizeImagPartFn := func(imagPart string) string {
			if imagPart == "-" {
				return "-1"
			}
			if imagPart == "" {
				return "1"
			}
			if imagPart == "+" {
				return "+1"
			}
			return imagPart
		}
		if !(arg[0] == '(' && arg[len(arg)-1] == ')') {
			res := &gosif_ComplexArgStruct{}
			if arg[len(arg)-1] == 'i' {
				res.I = arg[:len(arg)-1]
				res.I = normalizeImagPartFn(res.I)
				res.R = "0"
				return res, nil
			}
			res.R = arg
			res.I = "0"
			return res, nil
		}
		argPayload := arg[1 : len(arg)-1]
		generalFormatErrMsg := "it should contain real and imaginary parts separated by a comma (e.g. \"(1,2i)\")"
		sepPos := strings.Index(argPayload, ",")
		if sepPos == -1 {
			return nil, generateFormatErrMsgFn(generalFormatErrMsg)
		}
		if sepPos == len(argPayload)-1 {
			return nil, generateFormatErrMsgFn(generalFormatErrMsg)
		}
		first := argPayload[:sepPos]
		second := argPayload[sepPos+1:]
		multipleSeparatorsErrMsg := "multiple commas found"
		if sepPos := strings.Index(second, ","); sepPos != -1 {
			return nil, generateFormatErrMsgFn(multipleSeparatorsErrMsg)
		}
		first = strings.TrimSpace(first)
		second = strings.TrimSpace(second)
		if len(first) == 0 || len(second) == 0 {
			return nil, generateFormatErrMsgFn(generalFormatErrMsg)
		}
		isFirstImaginary := first[len(first)-1] == 'i'
		isSecondImaginary := second[len(second)-1] == 'i'
		if (isFirstImaginary && isSecondImaginary) || (!isFirstImaginary && !isSecondImaginary) {
			return nil, generateFormatErrMsgFn("it should contain one and only one real part and one and only one imaginary part")
		}
		res := &gosif_ComplexArgStruct{}
		if isFirstImaginary {
			res.R, res.I = second, first[:len(first)-1]
		} else {
			res.R, res.I = first, second[:len(second)-1]
		}
		res.I = normalizeImagPartFn(res.I)
		return res, nil
	}`,
}

type gosif_ComplexArgStruct struct {
	R string
	I string
}

func gosif_ParseArgAsComplex(arg string) (*gosif_ComplexArgStruct, error) {
	if len(arg) == 0 {
		return nil, fmt.Errorf("an empty string cannot be parsed as a complex number")
	}
	generateFormatErrMsgFn := func(reason string) error {
		return fmt.Errorf("\"%s\" cannot be parsed as a complex number: %s", arg, reason)
	}
	normalizeImagPartFn := func(imagPart string) string {
		if imagPart == "-" {
			return "-1"
		}
		if imagPart == "" {
			return "1"
		}
		if imagPart == "+" {
			return "+1"
		}
		return imagPart
	}
	if !(arg[0] == '(' && arg[len(arg)-1] == ')') {
		res := &gosif_ComplexArgStruct{}
		if arg[len(arg)-1] == 'i' {
			res.I = arg[:len(arg)-1]
			res.I = normalizeImagPartFn(res.I)
			res.R = "0"
			return res, nil
		}
		res.R = arg
		res.I = "0"
		return res, nil
	}
	argPayload := arg[1 : len(arg)-1]
	generalFormatErrMsg := "it should contain real and imaginary parts separated by a comma (e.g. \"(1,2i)\")"
	sepPos := strings.Index(argPayload, ",")
	if sepPos == -1 {
		return nil, generateFormatErrMsgFn(generalFormatErrMsg)
	}
	if sepPos == len(argPayload)-1 {
		return nil, generateFormatErrMsgFn(generalFormatErrMsg)
	}
	first := argPayload[:sepPos]
	second := argPayload[sepPos+1:]
	multipleSeparatorsErrMsg := "multiple commas found"
	if sepPos := strings.Index(second, ","); sepPos != -1 {
		return nil, generateFormatErrMsgFn(multipleSeparatorsErrMsg)
	}
	first = strings.TrimSpace(first)
	second = strings.TrimSpace(second)
	if len(first) == 0 || len(second) == 0 {
		return nil, generateFormatErrMsgFn(generalFormatErrMsg)
	}
	isFirstImaginary := first[len(first)-1] == 'i'
	isSecondImaginary := second[len(second)-1] == 'i'
	if (isFirstImaginary && isSecondImaginary) || (!isFirstImaginary && !isSecondImaginary) {
		return nil, generateFormatErrMsgFn("it should contain one and only one real part and one and only one imaginary part")
	}
	res := &gosif_ComplexArgStruct{}
	if isFirstImaginary {
		res.R, res.I = second, first[:len(first)-1]
	} else {
		res.R, res.I = first, second[:len(second)-1]
	}
	res.I = normalizeImagPartFn(res.I)
	return res, nil
}

var funcCheckRequiredFlags predefinedFunc = predefinedFunc{
	name: "funcCheckRequiredFlags",
	body: `
	func gosif_CheckRequiredFlags(requiredFlags map[string]bool) error {
		for flag, isPresent := range requiredFlags {
			if !isPresent {
				return fmt.Errorf("a required flag \"-%s\" was not passed", flag)
			}
		}
		return nil
	}`,
}

func gosif_CheckRequiredFlags(requiredFlags map[string]bool) error {
	for flag, isPresent := range requiredFlags {
		if !isPresent {
			return fmt.Errorf("a required flag \"-%s\" was not passed", flag)
		}
	}
	return nil
}
