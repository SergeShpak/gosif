package generator

import (
	"fmt"
	"testing"
)

func Test_gosif_ParseArgAsComplex(t *testing.T) {
	type testCase struct {
		in          string
		expectedRes *gosif_ComplexArgStruct
		expectedErr error
	}
	cases := []testCase{
		{
			in: "123",
			expectedRes: &gosif_ComplexArgStruct{
				R: "123",
				I: "0",
			},
		},
		{
			in: "123i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "123",
				R: "0",
			},
		},
		{
			in: "123.456",
			expectedRes: &gosif_ComplexArgStruct{
				R: "123.456",
				I: "0",
			},
		},
		{
			in: "123.456i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "123.456",
				R: "0",
			},
		},
		{
			in: "+123.456",
			expectedRes: &gosif_ComplexArgStruct{
				R: "+123.456",
				I: "0",
			},
		},
		{
			in: "-123.456i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "-123.456",
				R: "0",
			},
		},
		{
			in: "+123.456i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "+123.456",
				R: "0",
			},
		},
		{
			in: "i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "1",
				R: "0",
			},
		},
		{
			in: "+i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "+1",
				R: "0",
			},
		},
		{
			in: "-i",
			expectedRes: &gosif_ComplexArgStruct{
				I: "-1",
				R: "0",
			},
		},
		{
			in: "some-string",
			expectedRes: &gosif_ComplexArgStruct{
				R: "some-string",
				I: "0",
			},
		},
		{
			in: "some-stringi",
			expectedRes: &gosif_ComplexArgStruct{
				I: "some-string",
				R: "0",
			},
		},
		{
			in: "(123.456, 654.321i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "123.456",
				I: "654.321",
			},
		},
		{
			in: "(123.456, -654.321i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "123.456",
				I: "-654.321",
			},
		},
		{
			in: "(-123.456, 654.321i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "-123.456",
				I: "654.321",
			},
		},
		{
			in: "(-123.456, -654.321i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "-123.456",
				I: "-654.321",
			},
		},
		{
			in: "(+123.456, +654.321i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "+123.456",
				I: "+654.321",
			},
		},
		{
			in: "(-123.456i, -654.321)",
			expectedRes: &gosif_ComplexArgStruct{
				I: "-123.456",
				R: "-654.321",
			},
		},
		{
			in: "(42, i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "42",
				I: "1",
			},
		},
		{
			in: "(-i, 42)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "42",
				I: "-1",
			},
		},
		{
			in: "(+i, 42)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "42",
				I: "+1",
			},
		},
		{
			in: "(42, 0i)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "42",
				I: "0",
			},
		},
		{
			in: "(   -123.456i    ,-654.321    )",
			expectedRes: &gosif_ComplexArgStruct{
				I: "-123.456",
				R: "-654.321",
			},
		},
		{
			in: "(   -123.456i    ,-654.321    )",
			expectedRes: &gosif_ComplexArgStruct{
				I: "-123.456",
				R: "-654.321",
			},
		},
		{
			in: "(real, imagi)",
			expectedRes: &gosif_ComplexArgStruct{
				R: "real",
				I: "imag",
			},
		},
		{
			in:          "",
			expectedErr: fmt.Errorf("an empty string cannot be parsed as a complex number"),
		},
		{
			in:          "(1)",
			expectedErr: fmt.Errorf("\"(1)\" cannot be parsed as a complex number: it should contain real and imaginary parts separated by a comma (e.g. \"(1,2i)\")"),
		},
		{
			in:          "(1,,2i)",
			expectedErr: fmt.Errorf("\"(1,,2i)\" cannot be parsed as a complex number: multiple commas found"),
		},
		{
			in:          "(   ,    )",
			expectedErr: fmt.Errorf("\"(   ,    )\" cannot be parsed as a complex number: it should contain real and imaginary parts separated by a comma (e.g. \"(1,2i)\")"),
		},
		{
			in:          "(123.456, 654.321)",
			expectedErr: fmt.Errorf("\"(123.456, 654.321)\" cannot be parsed as a complex number: it should contain one and only one real part and one and only one imaginary part"),
		},
		{
			in:          "(123.456i, 654.321i)",
			expectedErr: fmt.Errorf("\"(123.456i, 654.321i)\" cannot be parsed as a complex number: it should contain one and only one real part and one and only one imaginary part"),
		},
	}

	for i, tc := range cases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			t.Parallel()
			t.Logf("testing input %s", tc.in)
			actualRes, actualErr := gosif_ParseArgAsComplex(tc.in)
			if err := checkErrors(tc.expectedErr, actualErr); err != nil {
				t.Fatal(err)
			}
			if tc.expectedRes == nil && actualRes == nil {
				return
			}
			if tc.expectedRes == nil {
				t.Fatalf("testing input %s: expected a nil result, got %v", tc.in, *actualRes)
			}
			if actualRes == nil {
				t.Fatalf("testing input %s: expected result %v, got a nil result", tc.in, tc.expectedRes)
			}
			if *tc.expectedRes != *actualRes {
				t.Fatalf("testing input %s: expected result %v, got %v", tc.in, *tc.expectedRes, *actualRes)
			}
		})
	}
}

func Test_gosif_CheckRequiredflags(t *testing.T) {
	cases := []struct {
		in          map[string]bool
		expectedErr error
	}{
		{
			in: map[string]bool{
				"a": true,
				"b": true,
			},
		},
		{
			in: map[string]bool{
				"a": true,
				"b": false,
			},
			expectedErr: fmt.Errorf("a required flag \"-b\" was not passed"),
		},
		{
			in: map[string]bool{},
		},
		{
			in: nil,
		},
	}
	for i, tc := range cases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			t.Parallel()
			err := gosif_CheckRequiredFlags(tc.in)
			if err := checkErrors(tc.expectedErr, err); err != nil {
				t.Fatal(err)
			}
		})
	}
}
