package generator

import (
	"fmt"
	"testing"
)

func TestUtilExtractFlag(t *testing.T) {
	cases := []struct {
		in          string
		expected    string
		expectedErr error
	}{
		{
			in:       "-flag",
			expected: "flag",
		},
		{
			in:       "--flag",
			expected: "flag",
		},
		{
			in:       "-\"-flag",
			expected: "\"-flag",
		},
		{
			in:       "---flag",
			expected: "-flag",
		},
		{
			in:       "--",
			expected: "--",
		},
		{
			in:       "---",
			expected: "-",
		},
		{
			in:       "----",
			expected: "--",
		},
		{
			in:          "",
			expectedErr: fmt.Errorf("internal error: expected a flag, got an empty string"),
		},
		{
			in:          "-",
			expectedErr: fmt.Errorf("passed flag \"-\" is treated as empty and empty flags are not allowed"),
		},
		{
			in:          "\"-flag",
			expectedErr: fmt.Errorf("expected a flag (e.g. --flag), got an argument \"\"-flag\""),
		},
		{
			in:          "flag",
			expectedErr: fmt.Errorf("expected a flag (e.g. --flag), got an argument \"flag\""),
		},
	}
	for i, tc := range cases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			t.Parallel()
			actual, err := gosif_UtilExtractFlag(tc.in)
			if err := checkErrors(tc.expectedErr, err); err != nil {
				t.Fatal(err)
			}
			if actual != tc.expected {
				t.Fatalf("actual flag %s and expected flag %s are not equal", actual, tc.expected)
			}
		})
	}
}

func checkErrors(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if expectedErr == nil || actualErr == nil || expectedErr.Error() != actualErr.Error() {
		return fmt.Errorf("expected error %v, but got %v", expectedErr, actualErr)
	}
	return nil
}

func TestParseArgs(t *testing.T) {
	type inArg struct {
		args      []string
		funcFlags map[string]struct{}
	}
	type outArg struct {
		parsedFlags map[string]gosif_ReadFlag
		err         error
	}
	okCases := []struct {
		in       inArg
		expected map[string]gosif_ReadFlag
	}{
		{
			in: inArg{
				args: []string{"-a", "aArg1", "aArg2", "-b", "bArg1", "--c", "\"-a\"", "\\\"-b\\\"", "-d", "\"\"", "-e"},
				funcFlags: map[string]struct{}{
					"a": {},
					"b": {},
					"c": {},
					"d": {},
					"e": {},
				},
			},
			expected: map[string]gosif_ReadFlag{
				"a": {
					PassedFlag: "-a",
					Args:       []string{"aArg1", "aArg2"},
				},
				"b": {
					PassedFlag: "-b",
					Args:       []string{"bArg1"},
				},
				"c": {
					PassedFlag: "--c",
					Args:       []string{"-a", "\"-b\""},
				},
				"d": {
					PassedFlag: "-d",
					Args:       []string{""},
				},
				"e": {
					PassedFlag: "-e",
					Args:       []string{},
				},
			},
		},
		{
			in: inArg{
				args: []string{"-a", "aArg1", "-b", "-b"},
				funcFlags: map[string]struct{}{
					"a": {},
				},
			},
			expected: map[string]gosif_ReadFlag{
				"a": {
					PassedFlag: "-a",
					Args:       []string{"aArg1", "-b", "-b"},
				},
			},
		},
		{
			in: inArg{
				args: []string{"-a", "aArg", "-"},
				funcFlags: map[string]struct{}{
					"a": {},
				},
			},
			expected: map[string]gosif_ReadFlag{
				"a": {
					PassedFlag: "-a",
					Args:       []string{"aArg", "-"},
				},
			},
		},
		{
			in: inArg{
				args: []string{"-a", "aArg", "---"},
				funcFlags: map[string]struct{}{
					"a": {},
				},
			},
			expected: map[string]gosif_ReadFlag{
				"a": {
					PassedFlag: "-a",
					Args:       []string{"aArg", "---"},
				},
			},
		},
		{
			in: inArg{
				args: []string{},
				funcFlags: map[string]struct{}{
					"a": {},
					"b": {},
					"c": {},
				},
			},
			expected: map[string]gosif_ReadFlag{},
		},
		{
			in: inArg{
				args: nil,
				funcFlags: map[string]struct{}{
					"a": {},
					"b": {},
					"c": {},
				},
			},
			expected: map[string]gosif_ReadFlag{},
		},
		{
			in: inArg{
				args: []string{"--a", "aArg1", "-b", "bArg1", "--b", "bArg2", "-a", "aArg3", "--b", "bArg3", "-c", "cArg1", "-c", "cArg2", "-d", "dArg1", "--d"},
				funcFlags: map[string]struct{}{
					"a": {},
					"b": {},
					"c": {},
					"d": {},
				},
			},
			expected: map[string]gosif_ReadFlag{
				"a": {
					PassedFlag: "-a",
					Args:       []string{"aArg3"},
				},
				"b": {
					PassedFlag: "--b",
					Args:       []string{"bArg3"},
				},
				"c": {
					PassedFlag: "-c",
					Args:       []string{"cArg2"},
				},
				"d": {
					PassedFlag: "--d",
					Args:       []string{},
				},
			},
		},
	}
	eqMapStringParsedFlag := func(actual map[string]gosif_ReadFlag, expected map[string]gosif_ReadFlag) error {
		if actual == nil && expected == nil {
			return nil
		}
		if actual == nil {
			return fmt.Errorf("actual map is nil, but expected map %v is not nil", expected)
		}
		if expected == nil {
			return fmt.Errorf("actual map %v is not nil, but expected map is nil", actual)
		}
		if len(actual) != len(expected) {
			return fmt.Errorf("actual map %v (length %d) and expected map %v (length %d) are of different length", actual, len(actual), expected, len(expected))
		}
		for k, expectedVal := range expected {
			actualVal, ok := actual[k]
			if !ok {
				return fmt.Errorf("actual map %v and expected map %v are different: actual map does not contain key %s", actual, expected, k)
			}
			if actualVal.PassedFlag != expectedVal.PassedFlag {
				return fmt.Errorf("actual value %v and expected value %v associated with key %s are different: actual passed flag is %s, expected passed flag is %s", actualVal, expectedVal, k, actualVal.PassedFlag, expectedVal.PassedFlag)
			}
			if err := eqStrSlices(actualVal.Args, expectedVal.Args); err != nil {
				return fmt.Errorf("actual value %v and expected value %v associated with key %s are different: %v", actualVal, expectedVal, k, err)
			}
		}
		return nil
	}
	for i, tc := range okCases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("ok test case #%d", i), func(t *testing.T) {
			t.Parallel()
			actual, err := gosif_ReadArgs(tc.in.args, tc.in.funcFlags)
			if err != nil {
				t.Fatal(err)
			}
			if err := eqMapStringParsedFlag(actual, tc.expected); err != nil {
				t.Fatal(err)
			}
		})
	}
	errorCases := []struct {
		in       inArg
		expected error
	}{
		{
			in: inArg{
				args:      []string{"-a", "aArg", "-b", "bArg", "-c", "cArg"},
				funcFlags: nil,
			},
			expected: fmt.Errorf("internal error: function flags map cannot be empty"),
		},
		{
			in: inArg{
				args:      []string{"-a", "aArg", "-b", "bArg", "-c", "cArg"},
				funcFlags: map[string]struct{}{},
			},
			expected: fmt.Errorf("internal error: function flags map cannot be empty"),
		},
		{
			in: inArg{
				args: []string{"-a", "aArg", "--"},
				funcFlags: map[string]struct{}{
					"a": {},
				},
			},
			expected: fmt.Errorf("\"--\" found, but ending command options and passing positional arguments is NYI"),
		},
	}

	for i, tc := range errorCases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("error test case #%d", i), func(t *testing.T) {
			t.Parallel()
			actual, err := gosif_ReadArgs(tc.in.args, tc.in.funcFlags)
			if err == nil || actual != nil {
				t.Fatalf("expected an error, but parsed flags %v and an error %v were returned", actual, err)
			}
			if err.Error() != tc.expected.Error() {
				t.Fatalf("actual and expected errors are not equal:\n\tactual: %v\n\texpected: %v", err, tc.expected)
			}
		})
	}
}

func TestUtilExtractArg(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{
			in:       "arg",
			expected: "arg",
		},
		{
			in:       "-arg",
			expected: "-arg",
		},
		{
			in:       "\"-arg\"",
			expected: "-arg",
		},
		{
			in:       "\"\"",
			expected: "",
		},
		{
			in:       "",
			expected: "",
		},
		{
			in:       "\\\"-arg\\\"",
			expected: "\"-arg\"",
		},
		{
			in:       "\"",
			expected: "\"",
		},
		{
			in:       "\\\"\\\"",
			expected: "\"\"",
		},
		{
			in:       "\\\"\"\"\\\"",
			expected: "\"\"\"\"",
		},
	}

	for i, tc := range cases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			t.Parallel()
			actual := gosif_UtilExtractArg(tc.in)
			if actual != tc.expected {
				t.Fatalf("actual argument %s and expected argument %s are not equal", actual, tc.expected)
			}
		})
	}
}

func TestUtilReadFlagArgs(t *testing.T) {
	type inArg struct {
		args      []string
		funcFlags map[string]struct{}
	}
	type okCaseNoFlagsBase struct {
		inArgs   []string
		expected []string
	}
	type okCase struct {
		in       inArg
		expected []string
	}
	generateOkCasesNoFlags := func(bases []okCaseNoFlagsBase) []okCase {
		cases := make([]okCase, len(bases)*2)
		for i, b := range bases {
			cases[i*2] = okCase{
				in: inArg{
					args:      b.inArgs,
					funcFlags: map[string]struct{}{},
				},
				expected: b.expected,
			}
			cases[i*2+1] = okCase{
				in: inArg{
					args: b.inArgs,
					funcFlags: map[string]struct{}{
						"notPassedFlag1": {},
						"notPassedFlag2": {},
					},
				},
				expected: b.expected,
			}
		}
		return cases
	}
	okCasesNoFlagsBases := []okCaseNoFlagsBase{
		{
			inArgs:   []string{"a", "b", "-c", "--d"},
			expected: []string{"a", "b", "-c", "--d"},
		},
		{
			inArgs:   []string{"a", "b", "\"-c\"", "d"},
			expected: []string{"a", "b", "-c", "d"},
		},
		{
			inArgs:   []string{"a", "b", "\\\"-c\\\"", "d"},
			expected: []string{"a", "b", "\"-c\"", "d"},
		},
		{
			inArgs:   []string{},
			expected: []string{},
		},
		{
			inArgs:   nil,
			expected: []string{},
		},
	}
	runTest := func(t *testing.T, tc okCase) {
		t.Logf("in args: %v, in func flags: %v", tc.in.args, tc.in.funcFlags)
		actual, err := gosif_UtilReadFlagArgs(tc.in.args, tc.in.funcFlags)
		if err != nil {
			t.Fatal(err)
		}
		if err := eqStrSlices(actual, tc.expected); err != nil {
			t.Fatal(err)
		}
	}
	okCasesNoFlags := generateOkCasesNoFlags(okCasesNoFlagsBases)
	for i, tc := range okCasesNoFlags {
		i, tc := i, tc
		t.Run(fmt.Sprintf("ok no flags test case #%d", i), func(t *testing.T) {
			t.Parallel()
			runTest(t, tc)
		})
	}
	okCases := []okCase{
		{
			in: inArg{
				args: []string{"a", "b", "c", "d"},
				funcFlags: map[string]struct{}{
					"c": {},
				},
			},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			in: inArg{
				args: []string{"a", "-b", "c", "d"},
				funcFlags: map[string]struct{}{
					"b": {},
				},
			},
			expected: []string{"a"},
		},
		{
			in: inArg{
				args: []string{"a", "b", "-c", "d"},
				funcFlags: map[string]struct{}{
					"b": {},
				},
			},
			expected: []string{"a", "b", "-c", "d"},
		},
	}
	for i, tc := range okCases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("ok test case #%d", i), func(t *testing.T) {
			t.Parallel()
			runTest(t, tc)
		})
	}
}

func eqStrSlices(actual []string, expected []string) error {
	if actual == nil && expected == nil {
		return nil
	}
	if actual == nil {
		return fmt.Errorf("actual slice is nil, but expected slice %v is not nil", expected)
	}
	if expected == nil {
		return fmt.Errorf("actual slice %v is not nil, but expected slice is nil", actual)
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("actual slice %v (length %d) and expected slice %v (length %d) are of diffrent length", actual, len(actual), expected, len(expected))
	}
	for i, actualEl := range actual {
		expectedEl := expected[i]
		if actualEl != expectedEl {
			return fmt.Errorf("actual slice %v and expected slice %v are different: elements %s and %s on index %d are not equal", actual, expected, actualEl, expectedEl, i)
		}
	}
	return nil
}
