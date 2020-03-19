//+build integration_tests

package main_func

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/SergeyShpak/gosif/tests/utils"
)

const outBin = "test_bin"
const outDir = "test"

var flagPrefixes []string = []string{"-", "--"}

type arg struct {
	flag string
	val  string
}

type sliceArg struct {
	flag string
	vals []string
}

func TestScripts(t *testing.T) {
	if err := utils.Setup(outBin, outDir, true); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cleanup := func() {
		utils.RemoveArtifacts(outBin, outDir)
	}
	t.Cleanup(cleanup)
	t.Run("Test scripts", func(t *testing.T) {
		scriptsTestData := []struct {
			scriptName string
			args       []arg
		}{
			{
				scriptName: "IntScript",
				args:       []arg{{flag: "n", val: "42"}, {flag: "n8", val: "-128"}, {flag: "n16", val: "32767"}, {flag: "n32", val: "-2147483648"}, {flag: "n64", val: "9223372036854775807"}},
			},
			{
				scriptName: "UintScript",
				args:       []arg{{flag: "n", val: "42"}, {flag: "n8", val: "255"}, {flag: "n16", val: "65535"}, {flag: "n32", val: "4294967295"}, {flag: "n64", val: "18446744073709551615"}},
			},
			{
				scriptName: "FloatScript",
				args:       []arg{{flag: "n32", val: "123.456"}, {flag: "n64", val: "-123.456"}},
			},
			{
				scriptName: "StringScript",
				args:       []arg{{flag: "s", val: "some string"}},
			},
			{
				scriptName: "ByteScript",
				args:       []arg{{flag: "b", val: "42"}},
			},
			{
				scriptName: "RuneScript",
				args:       []arg{{flag: "r", val: "⌘"}},
			},
			{
				scriptName: "ComplexScript",
				args:       []arg{{flag: "c64", val: "(123.456,456.123i)"}, {flag: "c128", val: "(-123.456,-456.123i)"}},
			},
			{
				scriptName: "ComplexScript",
				args:       []arg{{flag: "c64", val: "(0.000,0.000i)"}, {flag: "c128", val: "(0.000,0.000i)"}},
			},
			{
				scriptName: "ErrorScript",
				args:       []arg{{flag: "e", val: "an error message"}},
			},
		}
		cases := make([]utils.TestCase, 0)
		for _, s := range scriptsTestData {
			casesWithPointers := generateOKCases(s.scriptName, s.args, true)
			cases = append(cases, casesWithPointers...)
			casesWithoutPointers := append(cases, casesWithPointers...)
			cases = append(cases, casesWithoutPointers...)
		}
		cases = append(cases, []utils.TestCase{
			{
				ScriptName:  "ComplexScript",
				Args:        []string{"-c64", "42", "-c128", "-42i"},
				ExpectedOut: "c64: (42.000,0.000i), c128: (0.000,-42.000i)\nc64p: nil, c128p: nil",
			},
		}...)

		errorCases := []utils.TestCase{
			{
				ExpectedErr: fmt.Errorf("[ERR]: no function name passed"),
			},
			{
				ScriptName:  "SimpleScript",
				ExpectedErr: fmt.Errorf("[ERR]: a required flag \"-arg\" was not passed"),
			},
			{
				ScriptName:  "SimpleScript",
				Args:        []string{"arg"},
				ExpectedErr: fmt.Errorf("[ERR]: an error occurred during the flag \"arg\" extraction: expected a flag (e.g. --flag), got an argument \"arg\""),
			},
			{
				ScriptName:  "SimpleScript",
				Args:        []string{"-"},
				ExpectedErr: fmt.Errorf("[ERR]: an error occurred during the flag \"-\" extraction: passed flag \"-\" is treated as empty and empty flags are not allowed"),
			},
			{
				ScriptName:  "SimpleScript",
				Args:        []string{"--"},
				ExpectedErr: fmt.Errorf("[ERR]: \"--\" found, but ending command options and passing positional arguments is NYI"),
			},
			{
				ScriptName:  "SimpleScript",
				Args:        []string{"--n", "123"},
				ExpectedErr: fmt.Errorf("[ERR]: an unexpected flag \"--n\" found"),
			},
			{
				ScriptName:  "SimpleScript",
				Args:        []string{"-n", "123"},
				ExpectedErr: fmt.Errorf("[ERR]: an unexpected flag \"-n\" found"),
			},
			{
				ScriptName:  "IntScript",
				Args:        []string{"-n", "123", "-n8", "8", "-n16", "not a number", "-n32", "32", "-n64", "64"},
				ExpectedErr: fmt.Errorf("[ERR]: cast failed: failed to cast not a number to int16: strconv.ParseInt: parsing \"not a number\": invalid syntax"),
			},
			{
				ScriptName:  "IntScript",
				Args:        []string{"-n", "123", "-n8", "256", "-n16", "16", "-n32", "32", "-n64", "64"},
				ExpectedErr: fmt.Errorf("[ERR]: cast failed: failed to cast 256 to int8: strconv.ParseInt: parsing \"256\": value out of range"),
			},
			{
				ScriptName:  "IntScript",
				Args:        []string{"-n", "123", "-n8", "-n16", "16", "-n32", "32", "-n64", "64"},
				ExpectedErr: fmt.Errorf("[ERR]: could not get the argument passed to the flag \"-n8\": no arguments passed"),
			},
			{
				ScriptName:  "IntScript",
				Args:        []string{"-n", "123", "-n8", "8", "9", "-n16", "16", "-n32", "32", "-n64", "64"},
				ExpectedErr: fmt.Errorf("[ERR]: could not get the argument passed to the flag \"-n8\": expected a single argument, got 2 ([8 9])"),
			},
			// casting errors
			{
				ScriptName:  "ComplexScript",
				Args:        []string{"-c64", "(real,imagi)", "-c128", "(1,i)"},
				ExpectedErr: fmt.Errorf("[ERR]: cast failed: failed to cast \"real\" to float32: strconv.ParseFloat: parsing \"real\": invalid syntax"),
			},
		}
		unknownScripts := []string{"StructScript", "unexportedScript", "InexistentScript"}
		for _, s := range unknownScripts {
			errorCases = append(errorCases, utils.TestCase{
				ScriptName:  s,
				ExpectedErr: fmt.Errorf("[ERR]: unknown function %s", s),
			})
		}
		cases = append(cases, errorCases...)
		for i, tc := range cases {
			i, tc := i, tc
			t.Run(fmt.Sprintf("test #%d for script %s", i, tc.ScriptName), func(t *testing.T) {
				t.Parallel()
				t.Logf("scripts arguments: %v", tc.Args)
				out, err := utils.RunScript(path.Join(outDir, outBin), tc.ScriptName, tc.Args)
				if err := utils.CheckRunScriptResult(&tc, out, err); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("Test bool script", func(t *testing.T) {
		trueBoolArgs := []string{"true", "t", "T", "TRUE", "tRUe"}
		falseBoolArgs := []string{"false", "f", "FALSE", "F", "fAlSe"}
		boolTestCases := make([]utils.TestCase, 0)
		generateOKBoolCasesFn := func(arg string, expectedBool bool) []utils.TestCase {
			cases := make([]utils.TestCase, 0, 3*len(flagPrefixes))
			baseCase := utils.TestCase{
				ScriptName: "BoolScript",
			}
			for _, flagPrefix := range flagPrefixes {
				caseWithoutPointer := baseCase
				caseWithoutPointer.Args = []string{
					fmt.Sprintf("%sb", flagPrefix),
					arg,
				}
				caseWithoutPointer.ExpectedOut = fmt.Sprintf("b: %t\nbp: nil", expectedBool)
				cases = append(cases, caseWithoutPointer)

				caseWithPointer := baseCase
				caseWithPointer.Args = []string{
					fmt.Sprintf("%sb", flagPrefix),
					arg,
					fmt.Sprintf("%sbp", flagPrefix),
					arg,
				}
				caseWithPointer.ExpectedOut = fmt.Sprintf("b: %t\nbp: %t", expectedBool, expectedBool)
				cases = append(cases, caseWithPointer)

				caseWithOnlyPointer := baseCase
				caseWithOnlyPointer.Args = []string{
					fmt.Sprintf("%sbp", flagPrefix),
					arg,
				}
				caseWithOnlyPointer.ExpectedOut = fmt.Sprintf("b: false\nbp: %t", expectedBool)
				cases = append(cases, caseWithOnlyPointer)
			}
			return cases
		}
		for _, tArg := range trueBoolArgs {
			cases := generateOKBoolCasesFn(tArg, true)
			boolTestCases = append(boolTestCases, cases...)
		}
		for _, fArg := range falseBoolArgs {
			cases := generateOKBoolCasesFn(fArg, false)
			boolTestCases = append(boolTestCases, cases...)
		}
		boolTestCases = append(boolTestCases, []utils.TestCase{
			{
				ScriptName:  "BoolScript",
				Args:        []string{"-b"},
				ExpectedOut: "b: true\nbp: nil",
			},
			{
				ScriptName:  "BoolScript",
				Args:        []string{},
				ExpectedOut: "b: false\nbp: nil",
			},
		}...)
		for i, tc := range boolTestCases {
			i, tc := i, tc
			t.Run(fmt.Sprintf("test #%d for script %s", i, tc.ScriptName), func(t *testing.T) {
				t.Parallel()
				t.Logf("scripts arguments: %v", tc.Args)
				out, err := utils.RunScript(path.Join(outDir, outBin), tc.ScriptName, tc.Args)
				if err := utils.CheckRunScriptResult(&tc, out, err); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("Test No args script", func(t *testing.T) {
		out, err := utils.RunScript(path.Join(outDir, outBin), "NoArgsScript", nil)
		if err != nil {
			t.Fatal(err)
		}
		expectedOut := "this script does not expect any args"
		if out != expectedOut {
			t.Fatalf("expected output %s, got %s", expectedOut, out)
		}
	})
	t.Run("Test slice scripts", func(t *testing.T) {
		type TestData struct {
			scriptName  string
			args        sliceArg
			expectedStr *string
		}
		scriptsTestData := []TestData{
			{
				scriptName: "StringSliceScript",
				args:       sliceArg{flag: "s", vals: []string{"Hello", "gosif", "!"}},
			},
			{
				scriptName: "StringSliceScript",
				args:       sliceArg{flag: "s", vals: []string{}},
			},
			{
				scriptName: "IntSliceScript",
				args:       sliceArg{flag: "n", vals: []string{"-9223372036854775808", "9223372036854775807", "42"}},
			},
			{
				scriptName: "BoolSliceScript",
				args:       sliceArg{flag: "b", vals: []string{}},
			},
			{
				scriptName:  "BoolSliceScript",
				args:        sliceArg{flag: "b", vals: []string{"t", "f", "true", "false", "TRUE", "FALSE", "TrUe", "FaLsE"}},
				expectedStr: strConstToPtr("true false true false true false true false"),
			},
		}
		cases := make([]utils.TestCase, 0)
		for _, td := range scriptsTestData {
			cases = append(cases, generateSliceOKCases(td.scriptName, td.args, td.expectedStr)...)
		}
		arraysTestData := []TestData{
			{
				scriptName: "StringArrScript",
				args:       sliceArg{flag: "s", vals: []string{"Hello", "gosif", "!"}},
			},
		}
		for _, td := range arraysTestData {
			cases = append(cases, generateArrOKCases(td.scriptName, td.args, td.expectedStr)...)
		}
		cases = append(cases, []utils.TestCase{
			{
				ScriptName:  "StringArrScript",
				Args:        []string{"--s", "1", "2", "--sp", "1", "2"},
				ExpectedErr: fmt.Errorf("[ERR]: flag --s: expected 3 arguments, but got 2 ([1 2])"),
			},
			{
				ScriptName:  "StringArrScript",
				Args:        []string{"--s", "1", "2", "3", "--sp", "1", "2"},
				ExpectedOut: "s: 1 2 3\nsp: 1 2 nil\nps: nil slice\npsp: nil slice",
			},
			{
				ScriptName:  "StringArrScript",
				Args:        []string{"--s", "1", "2", "3", "--sp", "1", "2", "3", "4"},
				ExpectedErr: fmt.Errorf("[ERR]: flag --sp: expected 3 arguments, but got 4 ([1 2 3 4])"),
			},
			{
				ScriptName:  "StringArrLengthOneScript",
				Args:        []string{"--s", "1", "2"},
				ExpectedErr: fmt.Errorf("[ERR]: flag --s: expected 1 argument, but got 2 ([1 2])"),
			},
		}...)
		for i, tc := range cases {
			i, tc := i, tc
			t.Run(fmt.Sprintf("test cases #%d", i), func(t *testing.T) {
				t.Parallel()
				t.Logf("testing script %s", tc.ScriptName)
				t.Logf("passed slice: %v", tc.Args)
				out, err := utils.RunScript(path.Join(outDir, outBin), tc.ScriptName, tc.Args)
				if err := utils.CheckRunScriptResult(&tc, out, err); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}

func generateExpectedOutStr(scriptArgs []arg, withPointers bool) string {
	directVals := make([]string, len(scriptArgs))
	for i, arg := range scriptArgs {
		directVals[i] = fmt.Sprintf("%s: %s", arg.flag, arg.val)
	}
	directValsStr := strings.Join(directVals, ", ")
	pointers := make([]string, len(scriptArgs))
	if withPointers {
		for i, arg := range scriptArgs {
			pointers[i] = fmt.Sprintf("%sp: %s", arg.flag, arg.val)
		}
	} else {
		for i, arg := range scriptArgs {
			pointers[i] = fmt.Sprintf("%sp: nil", arg.flag)
		}
	}
	pointersStr := strings.Join(pointers, ", ")
	outStr := fmt.Sprintf("%s\n%s", directValsStr, pointersStr)
	return outStr
}

func generateOKCases(scriptName string, scriptArgs []arg, withPointers bool) []utils.TestCase {
	cases := make([]utils.TestCase, 0)
	outStr := generateExpectedOutStr(scriptArgs, withPointers)
	for _, p := range flagPrefixes {
		args := make([]string, 0)
		for _, sa := range scriptArgs {
			args = append(args, fmt.Sprintf("%s%s", p, sa.flag))
			args = append(args, sa.val)
			if withPointers {
				args = append(args, fmt.Sprintf("%s%sp", p, sa.flag))
				args = append(args, sa.val)
			}
		}
		c := utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: outStr,
		}
		cases = append(cases, c)
	}
	return cases
}

func generateSliceOKCases(scriptName string, scriptArg sliceArg, expectedOut *string) []utils.TestCase {
	cases := make([]utils.TestCase, 0)
	var joinedVals string
	if expectedOut == nil {
		joinedVals = strings.Join(scriptArg.vals, " ")
	} else {
		joinedVals = *expectedOut
	}
	for _, p := range flagPrefixes {
		// all flags
		flags := []string{scriptArg.flag, scriptArg.flag + "p", "p" + scriptArg.flag, "p" + scriptArg.flag + "p"}
		args := make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut := generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, joinedVals, joinedVals)
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
		// all flags in different order
		flags = []string{scriptArg.flag + "p", scriptArg.flag, "p" + scriptArg.flag + "p", "p" + scriptArg.flag}
		args = make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut = generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, joinedVals, joinedVals)
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
		// empty first flag
		flags = []string{scriptArg.flag + "p", "p" + scriptArg.flag + "p", "p" + scriptArg.flag}
		args = []string{fmt.Sprintf("%s%s", p, scriptArg.flag)}
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut = generateExpectedOutStrForSlice(scriptArg.flag, "", joinedVals, joinedVals, joinedVals)
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
		// no pointers
		flags = []string{scriptArg.flag, scriptArg.flag + "p"}
		args = make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut = generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, "nil slice", "nil slice")
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
	}
	return cases
}

func generateArrOKCases(scriptName string, scriptArg sliceArg, expectedOut *string) []utils.TestCase {
	cases := make([]utils.TestCase, 0)
	var joinedVals string
	if expectedOut == nil {
		joinedVals = strings.Join(scriptArg.vals, " ")
	} else {
		joinedVals = *expectedOut
	}
	for _, p := range flagPrefixes {
		// all flags
		flags := []string{scriptArg.flag, scriptArg.flag + "p", "p" + scriptArg.flag, "p" + scriptArg.flag + "p"}
		args := make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut := generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, joinedVals, joinedVals)
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
		// all flags in different order
		flags = []string{scriptArg.flag + "p", scriptArg.flag, "p" + scriptArg.flag + "p", "p" + scriptArg.flag}
		args = make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut = generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, joinedVals, joinedVals)
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
		// no pointers
		flags = []string{scriptArg.flag, scriptArg.flag + "p"}
		args = make([]string, 0)
		for _, f := range flags {
			flagWithPrefix := fmt.Sprintf("%s%s", p, f)
			args = append(args, append([]string{flagWithPrefix}, scriptArg.vals...)...)
		}
		expectedOut = generateExpectedOutStrForSlice(scriptArg.flag, joinedVals, joinedVals, "nil slice", "nil slice")
		cases = append(cases, utils.TestCase{
			ScriptName:  scriptName,
			Args:        args,
			ExpectedOut: expectedOut,
		})
	}
	return cases
}

func generateExpectedOutStrForSlice(flag string, a string, ap string, pa string, pap string) string {
	out := fmt.Sprintf("%[1]s: %[2]s\n%[1]sp: %[3]s\np%[1]s: %[4]s\np%[1]sp: %[5]s", flag, a, ap, pa, pap)
	return out
}

func strConstToPtr(str string) *string {
	return &str
}
