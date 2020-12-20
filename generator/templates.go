package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/SergeyShpak/gosif/generator/types"
)

type tmplCastFunctionNameInput struct {
	Type string
}

var tmplCastFunctionName = template.Must(tmplRunScriptFuncName.New("CastFunctionName").
	Parse("gosif_Parse_{{.Type}}_Arg"))

var tmplCastFunctionString = template.Must(tmplCastFunctionPostfix.New("CastFunctionString").Parse(`
{{template "CastFunctionPrefix" .}}
val = arg
{{- template "CastFunctionPostfix" .}}`))

var tmplCastFunctionError = template.Must(tmplCastFunctionPostfix.New("CastFunctionError").Parse(`
{{template "CastFunctionPrefix" .}}
val = fmt.Errorf(arg)
{{- template "CastFunctionPostfix" .}}`))

var tmplCastFunctionByte = template.Must(tmplCastFunctionPostfix.New("CastFunctionByte").
	Parse(`{{template "CastFunctionPrefix" .}}
val = fmt.Errorf(arg)
{{- template "CastFunctionPostfix" .}}`))

var tmplCastFunctionRune = template.Must(tmplCastFunctionPostfix.New("CastFunctionRune").
	Parse(`{{template "CastFunctionPrefix" .}}
argRunes := []rune(arg)
if len(argRunes) > 1 {
	return val, fmt.Errorf("failed to cast %s to rune: %s contains %d runes", arg, arg, len(argRunes))
} else if len(argRunes) == 0 {
	return val, fmt.Errorf("failed to cast an empty string to rune")
}
val = argRunes[0]
{{- template "CastFunctionPostfix" .}}`))

type castFuncComplexInput castFuncSizedTypeInput

var tmplCastFunctionComplex = template.Must(tmplCastFunctionPostfix.New("CastfunctionComplex").
	Funcs(template.FuncMap{
		"getType": func(bitSize string) (res tmplCastFunctionNameInput, err error) {
			switch bitSize {
			case "64":
				res = tmplCastFunctionNameInput{
					Type: "float32",
				}
			case "128":
				res = tmplCastFunctionNameInput{
					Type: "float64",
				}
			default:
				err = fmt.Errorf("internal error: unexpected bitSize (%s) in tmplCastFunctionComplex getType function", bitSize)
			}
			return
		},
	}).
	Parse(`{{template "CastFunctionPrefix" .}}
		complexArgs, err := gosif_ParseArgAsComplex(arg)
		if err != nil {
			return val, nil
		}
		{{- $type := getType .BitSize }}
		valR, err := {{template "CastFunctionName" $type}}(complexArgs.R)
		if err != nil {
			return val, err
		}
		valI, err := {{template "CastFunctionName" $type}}(complexArgs.I)
		if err != nil {
			return val, err
		}
		val = complex(valR, valI)
		{{- template "CastFunctionPostfix" .}}`))

var indirFuncHelpers = template.FuncMap{
	"Iterate": func(count int) []string {
		result := make([]string, count)
		for i := 0; i < count; i++ {
			result[i] = fmt.Sprintf("val%d := &val%d", i+1, i)
		}
		return result
	},
	"Stars": func(count int) string {
		return strings.Repeat("*", count)
	},
}

type tmplIndirFunctionNameInput tmplIndirFunctionInput

var tmplIndirFunctionName = template.Must(tmplCastFunctionName.New("IndirFunctionName").
	Parse(`gosif_{{.Type}}_Indir{{.IndirectionLevel}}`))

type tmplIndirFunctionInput struct {
	IndirectionLevel int
	Type             string
}

var tmplIndirFunction = template.Must(tmplIndirFunctionName.New("IndirFunction").
	Funcs(indirFuncHelpers).
	Parse(`func {{template "IndirFunctionName" .}}(val0 {{.Type}}) {{Stars .IndirectionLevel}}{{.Type}} {
	{{- range $val := Iterate .IndirectionLevel }}
		{{$val}}
	{{- end }}
	return val{{.IndirectionLevel}}
}`))

type tmplIndirArrFunctionNameInput tmplIndirArrFunctionInput

var tmplIndirArrFunctionName = template.Must(tmplCastFunctionName.New("IndirArrFunctionName").
	Funcs(template.FuncMap{
		"escapeElType": func(elType string) string {
			return strings.Replace(elType, "*", "_", -1)
		},
	}).
	Parse(`gosif_Arr{{if not .ArrInfo.IsSlice}}{{.ArrInfo.ArrayLength}}{{end}}{{escapeElType .ArrInfo.ElType}}_Indir{{.IndirectionLevel}}`))

type tmplIndirArrFunctionInput struct {
	ArrInfo          tmplArrayInfo
	IndirectionLevel int
}

var tmplIndirArrFunction = template.Must(tmplIndirFunctionName.New("IndirArrFunction").
	Funcs(indirFuncHelpers).
	Parse(`func {{template "IndirArrFunctionName" .}}(val0 [{{if not .ArrInfo.IsSlice}}{{.ArrInfo.ArrayLength}}{{end}}]{{.ArrInfo.ElType}}) {{Stars .IndirectionLevel}}[{{if not .ArrInfo.IsSlice}}{{.ArrInfo.ArrayLength}}{{end}}]{{.ArrInfo.ElType}} {
{{- range $val := Iterate .IndirectionLevel }}
	{{$val}}
{{- end }}
return val{{.IndirectionLevel}}
}`))

type tmplArgCastInput struct {
	IndirectionLevel int
	Type             string
	InArray          bool
}

var tmplArgCast = template.Must(tmplIndirFunctionName.New("ArgCast").
	Parse(`directVal, err := {{template "CastFunctionName" .}}(arg)
if err != nil {
	return nil, fmt.Errorf("cast failed: %v", err)
}
{{- if eq .IndirectionLevel 0}}
	val := directVal
{{- else }}
	val := {{template "IndirFunctionName" .}}(directVal)
{{- end }}`))

type tmplArrayInfo struct {
	IsSlice     bool
	ArrayLength int
	ElType      string
}

type tmplArrayLayerInput struct {
	ArrInfo              tmplArrayInfo
	Payload              string
	IndirectionLevel     int
	BaseIndirectionLevel int
}

var tmplArrayLayer = template.Must(tmplIndirArrFunctionName.New("ArrayLayer").
	Parse(`{{- if not .ArrInfo.IsSlice -}}
	{{- if eq .BaseIndirectionLevel 0 -}}
if len(parsedFlag.Args) != {{.ArrInfo.ArrayLength}} {
	return nil, fmt.Errorf("flag %s: expected {{.ArrInfo.ArrayLength}} argument{{if ne .ArrInfo.ArrayLength 1}}s{{end}}, but got %d (%v)", parsedFlag.PassedFlag, len(parsedFlag.Args), parsedFlag.Args)
}
	{{ else -}}
if len(parsedFlag.Args) > {{.ArrInfo.ArrayLength}} {
	return nil, fmt.Errorf("flag %s: expected {{.ArrInfo.ArrayLength}} argument{{if ne .ArrInfo.ArrayLength 1}}s{{end}}, but got %d (%v)", parsedFlag.PassedFlag, len(parsedFlag.Args), parsedFlag.Args)
}
{{ end -}}
{{ end -}}
{{if .ArrInfo.IsSlice -}}
directVal1 := make([]{{.ArrInfo.ElType}}, len(parsedFlag.Args))
{{else -}}
var directVal1 [{{.ArrInfo.ArrayLength}}]{{.ArrInfo.ElType}}
{{end -}}
for i, arg := range parsedFlag.Args {
	{{.Payload}}
	directVal1[i] = val
}
{{- if eq .IndirectionLevel 0 }}
	val1 := directVal1
{{- else }}
	val1 := {{template "IndirArrFunctionName" .}}(directVal1)
{{- end -}}`))

type tmplArgCastPostfixInput struct {
	FlagName  string
	IsPointer bool
	InArray   bool
	BaseType  string
}

var tmplArgCastPostfix = template.Must(template.New("ArgCastPostfix").
	Parse(`flags.{{.FlagName}} = {{if .IsPointer}}&{{end}}val{{if .InArray}}1{{end}}
{{- if not (or .IsPointer (and (eq .BaseType "bool") (not .InArray))) }}
	requiredFlags["{{.FlagName}}"] = true
{{- end -}}`))

type tmplArgCastPrefixInput struct {
	FlagName    string
	LayersCount int
	BaseType    string
}

var tmplArgCastPrefix = template.Must(template.New("ArgCastPrefix").
	Parse(`
case "{{.FlagName}}":
	{{- if eq .LayersCount 0 }}
	arg, err := gosif_Get{{- if eq .BaseType "bool" -}}Bool{{- else -}}Flag{{- end -}}Arg(parsedFlag.Args)
	if err != nil {
		return nil, fmt.Errorf("could not get the argument passed to the flag \"%s\": %v", parsedFlag.PassedFlag, err)
	}
	{{- end -}}`))

type tmplScriptsHelpFunctionInput struct {
	ScriptsNames []string
}

var tmplScriptsHelpFunction = template.Must(template.New("ScriptsHelpFunction").
	Parse(`
func gosif_ShowScriptsHelp(stream *os.File) {
	helpMsg := ` + "`" + `The following functions are available:
	{{- range $scriptName := .ScriptsNames }}
	{{$scriptName}}
	{{- end }}
{{- $exampleScriptName := "MyFunc" -}}
{{- if ne (len .ScriptsNames) 0 -}}
{{- $exampleScriptName = (index .ScriptsNames 0) -}}
{{- end -}}
To run a script pass its name as the first argument to the generated binary:
e.g. ./generated-binary {{$exampleScriptName}}
` + "`" + `
	fmt.Fprint(stream, helpMsg)	
}`))

type tmplFuncHelpFunctionInput struct {
	FunctionName  string
	Flags         []types.Flag
	RequiredFlags []types.Flag
}

var tmplFuncHelpFunction = template.Must(template.New("FuncHelpFunction").
	Parse(`
func gosif_Show{{.FunctionName}}Help(stream *os.File) {
	helpMsg := ` + "`" + `Function {{.FunctionName}}
	Required options:
		{{- range $flag := .RequiredFlags }}
		--{{$flag.Name | printf "%-10s"}}{{$flag.Type}}
		{{- end }}
	Available options:
		{{- range $flag := .Flags }}
		--{{$flag.Name | printf "%-10s"}}{{$flag.Type}}
		{{- end }}
` + "`" + `
	fmt.Fprint(stream, helpMsg)
}`))

type tmplParseFlagsFuncInput struct {
	FuncFlags     []types.Flag
	RequiredFlags []types.Flag
	Cases         []string
	FunctionName  string
}

var tmplParseFlagsFunc = template.Must(tmplRunScriptFuncName.New("ParseFlagsFunc").Parse(`
func {{template "ParseFlagsFuncName" .}}(args []string) (*{{template "FuncFlagsStructName" .}}, error) {
	funcFlags := map[string]struct{}{
		{{ range $flag := .FuncFlags -}}
		"{{$flag.Name}}": {},
		{{ end -}}
	}
	parsedArgs, err := gosif_ReadArgs(args, funcFlags)
	if err != nil {
		return nil, err
	}
	{{ if ne (len .RequiredFlags) 0 -}}
	requiredFlags := map[string]bool{
		{{ range $flag := .RequiredFlags -}}
		"{{$flag.Name}}": false,
		{{ end -}}
	}
	{{- end }}
	flags := &{{template "FuncFlagsStructName" .}}{}
	for name, parsedFlag := range parsedArgs {
		switch name {
			{{- range $case := .Cases}}{{$case}}{{end}}
		default:
			return nil, fmt.Errorf("internal error: a flag %s was expected, but no treating case had been generated", name)
		}
	}
	{{- if ne (len .RequiredFlags) 0 }}
	if err := gosif_CheckRequiredFlags(requiredFlags); err != nil {
		return nil, err
	}
	{{- end }}
	return flags, nil
}`))

var tmplFuncFlagsStructName = template.Must(template.New("FuncFlagsStructName").Parse(
	`gosif_{{- .FunctionName }}Flags`))
var tmplParseFlagsFuncName = template.Must(tmplFuncFlagsStructName.New("ParseFlagsFuncName").Parse(`gosif_Parse{{.FunctionName}}Flags`))
var tmplParseLoopName = template.Must(tmplParseFlagsFuncName.New("ParseLoopName").Parse(`ParseLoop`))
var tmplRunScriptFuncName = template.Must(tmplParseLoopName.New("RunScriptFuncName").Parse(`gosif_Run{{.FunctionName}}`))

type funcFlagStructureTmplInput struct {
	FunctionName string
	Flags        []types.Flag
}

var tmplFuncFlagsStruct = template.Must(tmplRunScriptFuncName.New("FuncFlagsStruct").Parse(`
type {{template "FuncFlagsStructName" .}} struct {
	{{ range $flag := .Flags -}}
		{{$flag.Name}} {{$flag.Type}}
	{{ end -}}
}`))

type tmplFlagCasePrefixInput struct {
	ArgName string
	ArgType string
}

var tmplFlagCasePrefix = template.Must(tmplCastFunctionName.New("FlagCasePrefix").Parse(`
case "{{.ArgName}}":`))

var tmplFlagCaseParsingErrCheck = template.Must(tmplFlagCasePrefix.New("FlagCaseParsingErrCheck").Parse(`
if err != nil {
	return nil, fmt.Errorf("flag %s: %v", parsedFlag.PassedFlag, err)
}`))

var tmplFlagCasePostfix = template.Must(tmplFlagCaseParsingErrCheck.New("FlagCasePostfix").Parse(`
if err != nil {
	return nil, fmt.Errorf("flag %s: %v", parsedFlag.PassedFlag, err)
}
flags.{{.ArgName}} = val
{{- if not (or .IsPointer (eq .CoreType "bool") ) }}
	requiredFlags["{{.ArgName}}"] = true
{{- end }}`))

type tmplFlagCaseInput struct {
	Prefix  string
	Body    string
	Postfix string
}

var tmplFlagCase = template.Must(template.New("FlagCase").
	Parse(`{{.Prefix}}
{{.Body}}
{{.Postfix}}`))

type simpleArgCaseTmplInput struct {
	ArgName   string
	ArgType   string
	CoreType  string
	IsPointer bool
}

var tmplBoolArgCase = template.Must(tmplFlagCasePostfix.New("BoolArgCase").
	Parse(`{{template "FlagCasePrefix" .}}
{{template "FlagCaseZeroOrSingleArgCheck" .}}
if len(parsedFlag.Args) == 0 {
	{{if .IsPointer}}*{{end}}val = true
} else {
	val, err = {{template "CastFunctionName" .}}(parsedFlag.Args[0])
}
{{- template "FlagCasePostfix" .}}`))

var tmplRuneArgCase = template.Must(tmplFlagCasePostfix.New("RuneArgCase").
	Parse(`{{template "FlagCasePrefix" .}}
argRunes := []rune(arg)
if len(argRunes) > 1 {
	return val, fmt.Errorf("failed to cast %s to rune: %s contains more than one rune", arg)
} else if len(argRunes) == 0 {
	return val, fmt.Errorf("failed to cast %s to rune: %s is an empty string", arg)
}
val, err = {{template "CastFunctionName" .}}(argRunes[0])
{{- template "FlagCasePostfix" .}}`))

type tmplStringArgCastInput struct {
	CastFunctionName string
	Ampersands       string
}

var tmplStringArgCast = template.Must(template.New("StringArgCast").
	Parse(`{{if .Ampersands}}directVal{{else}}val{{end}}, err := {{.CastFunctionName}}(arg)
if err != nil {
	return nil, fmt.Errorf("cast failed: %v", err)
}
{{if .Ampersands}}val := {{.Ampersands}}directVal{{end}}`))

type newSliceCaseInput struct {
	IsSlice       bool
	ArrLength     int
	ElemType      string
	PrevLayer     string
	ArrLayerIndex int
}

type tmplNewFlagCasePostfixInput struct {
	IsRequired         bool
	Ampersands         string
	ReturnValueVarName string
	ArgName            string
}

var tmplNewFlagCasePostfix = template.Must(tmplFlagCaseParsingErrCheck.New("NewFlagCasePostfix").
	Parse(`flags.{{.ArgName}} = {{.Ampersands}}{{.ReturnValueVarName}}
{{- if .IsRequired }}
	requiredFlags["{{.ArgName}}"] = true
{{- end }}`))

type castFunctionBaseInput tmplCastFunctionNameInput

var tmplCastFunctionPrefix = template.Must(tmplCastFunctionName.New("CastFunctionPrefix").Parse(`
func {{template "CastFunctionName" .}}(arg string) ({{.Type}}, error) {
	var val {{.Type}}
	var err error`))

var tmplCastFunctionPostfix = template.Must(tmplCastFunctionPrefix.New("CastFunctionPostfix").Parse(`
return val, err
}`))

var tmplCastFunctionBool = template.Must(tmplCastFunctionPostfix.New("CastFunctionBool").
	Parse(`{{template "CastFunctionPrefix" .}}
lowerParsedArg := strings.ToLower(arg)
if lowerParsedArg != "true" && lowerParsedArg != "t" && lowerParsedArg != "false" && lowerParsedArg != "f" {
	return val, fmt.Errorf("expected zero or one argument that must be any of [true, t, false, f] (case insensitive), got: %s", arg)
}
if lowerParsedArg == "true" || lowerParsedArg == "t" {
	val = true
}
{{- template "CastFunctionPostfix" .}}`))

type castFuncNumInput struct {
	castFunctionBaseInput
	BitSize string
	Signed  bool
}

// TODO: do we still need to check for the pointer in the error case?
var tmplCastFunctionNum = template.Must(tmplCastFunctionPostfix.New("CastFunctionNum").
	Parse(`{{template "CastFunctionPrefix" .}}
{{- $parseFunc := "ParseUint" -}}
{{- $varName := "valUint" -}}
{{- $castType := "uint" -}}
{{- if .Signed -}}
	{{- $parseFunc = "ParseInt" -}}
	{{- $varName = "valInt" -}}
	{{- $castType = "int" -}}
{{- end }}
{{$varName}}, err := strconv.{{$parseFunc}}(arg, 0, {{.BitSize}})
{{- $castExpr := $varName -}}
{{- if ne .BitSize "0" -}}
	{{- $castType = printf "%s%s" $castType .BitSize -}}
{{- end -}}
{{- if ne .BitSize "64" -}}
	{{- $castExpr = printf "(%s)(%s)" $castType $varName -}}
{{- end }}
if err != nil {
	return val, fmt.Errorf("failed to cast %s to {{$castType}}: %v", arg, err)
}
val = {{$castExpr}}
{{- template "CastFunctionPostfix" .}}`))

type castFuncSizedTypeInput struct {
	castFunctionBaseInput
	BitSize string
}

type castFuncFloatInput castFuncSizedTypeInput

var tmplCastFunctionFloat = template.Must(tmplCastFunctionPostfix.New("CastFunctionFloat").
	Parse(`{{template "CastFunctionPrefix" .}}
valFloat, err := strconv.ParseFloat(arg, {{.BitSize}})
if err != nil {
	return val, fmt.Errorf("failed to cast \"%s\" to float{{.BitSize}}: %v", arg, err)
}
val = {{if eq .BitSize "32"}}(float32)(valFloat){{else}}valFloat{{end}}
{{- template "CastFunctionPostfix" .}}`))

type runScriptFuncTmplInput funcFlagStructureTmplInput

var tmplRunScriptFunc = template.Must(tmplRunScriptFuncName.New("RunScriptFunc").Parse(`
func {{template "RunScriptFuncName" .}}(flags *{{template "FuncFlagsStructName" .}}) {
	{{.FunctionName}}({{range $i, $flag := .Flags}}{{if $i}},{{end}}flags.{{$flag.Name}}{{end}})
}`))

type mainFuncScriptCaseTmplInput struct {
	FunctionName string
}

var tmplMainFuncScriptCase = template.Must(tmplRunScriptFuncName.New("MainFuncScriptCase").Parse(`
case "{{.FunctionName}}":
	if len(os.Args) == 3 && os.Args[2] == "help" {
		gosif_Show{{.FunctionName}}Help(os.Stdout)
		return
	}
	flags, err := {{template "ParseFlagsFuncName" .}}(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR]: %v\n", err)
		gosif_Show{{- .FunctionName }}Help(os.Stderr)
		os.Exit(1)
	}
	{{template "RunScriptFuncName" .}}(flags)
	os.Exit(0)`))

var tmplMainFuncNoArgsScriptCase = template.Must(template.New("MainFuncScriptCase").Parse(`
case "{{.FunctionName}}":
	{{.FunctionName}}()
	os.Exit(0)`))

type mainFuncTmplInput struct {
	Cases   []string
	HasMain bool
}

var tmplMainFunc = template.Must(tmplRunScriptFuncName.New("MainFunc").Parse(`
func {{if .HasMain}}gosif{{else}}main{{end}}() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "[ERR]: no function name passed\n")
		gosif_ShowScriptsHelp(os.Stderr)
		os.Exit(1)
	}
	if len(os.Args) == 2 && os.Args[1] == "help" {
		gosif_ShowScriptsHelp(os.Stdout)
		return
	}
	switch os.Args[1] {
		{{- range $case := .Cases}}{{$case}}{{end}}
	default:
		fmt.Fprintf(os.Stderr, "[ERR]: unknown function %s\n", os.Args[1])
		gosif_ShowScriptsHelp(os.Stderr)
		os.Exit(1)
	}
}`))

type tmplFullFileInput struct {
	MainFunc        string
	Imports         []string
	Out             string
	CastFuncs       map[string]string
	IndirFuncs      map[string]string
	PredefinedFuncs map[string]string
}

var tmplFullFile = template.Must(template.New("FullFile").Parse(`
package main

import (
	"fmt"
	"os"
	{{ range $import := .Imports -}}
	"{{$import}}"
	{{ end -}}
)

{{.MainFunc}}

{{.Out}}

{{- range $castFunc := .CastFuncs }}
{{ $castFunc }}
{{- end }}

{{- range $indirFunc := .IndirFuncs }}
{{ $indirFunc }}
{{- end }}

{{- range $predefinedFunc := .PredefinedFuncs }}
{{ $predefinedFunc }}
{{- end }}`))
