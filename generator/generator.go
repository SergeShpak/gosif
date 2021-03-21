package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SergeyShpak/gosif/generator/trie"

	"github.com/SergeyShpak/gosif/generator/types"
	"github.com/SergeyShpak/gosif/parser"
)

const outFileName = "main.gen.go"

func GenerateScriptsForDir(dir string) error {
	if err := removePreviousOutput(dir); err != nil {
		return fmt.Errorf("failed to remove a previously generated file: %v", err)
	}
	packages, err := parser.ParsePackagesFunctions(dir)
	if err != nil {
		return err
	}
	mainPkg := packages.MainPackage
	if mainPkg == nil {
		return fmt.Errorf("main package was not found in %s", dir)
	}
	if len(mainPkg.Functions) == 0 {
		log.Println("gosif did not find any functions to process, nothing to generate")
		return nil
	}
	// TODO: refactor into several methods
	out, err := createMain(mainPkg)
	if err != nil {
		return err
	}
	helpFuncs, err := generateHelpFunctions(mainPkg.Functions)
	if err != nil {
		return err
	}
	out += helpFuncs

	if err := writeToFile(filepath.Join(dir, outFileName), out); err != nil {
		return err
	}
	return nil
}

type FuncForGenerator struct {
	ParsedFunc     *parser.PkgFunc
	OptionalParams []*FuncParamData
	RequiredParams []*FuncParamData
	Imports        map[string]struct{}
}

func extractDataFromParsedFunction(fn *parser.PkgFunc) (*FuncForGenerator, error) {
	data := &FuncForGenerator{
		ParsedFunc:     fn,
		OptionalParams: make([]*FuncParamData, 0),
		RequiredParams: make([]*FuncParamData, 0),
		Imports:        make(map[string]struct{}),
	}
	for i, param := range fn.Parameters {
		paramData, err := extractDataFromFuncParam(param)
		if err != nil {
			return nil, fmt.Errorf("failed to analyse parameters #%d \"%s\": %v", i, param.Name, err)
		}
		if paramData.IsOptional {
			data.OptionalParams = append(data.OptionalParams, paramData)
		} else {
			data.RequiredParams = append(data.RequiredParams, paramData)
		}
		for _, imp := range paramData.Imports {
			if _, ok := data.Imports[imp]; !ok {
				data.Imports[imp] = struct{}{}
			}
		}
	}
	allFlags := make([]*types.Flag, 0, len(fn.Parameters))
	for _, param := range data.OptionalParams {
		allFlags = append(allFlags, param.Flag)
	}
	for _, param := range data.RequiredParams {
		allFlags = append(allFlags, param.Flag)
	}
	if err := generateShortFlagsNames(allFlags); err != nil {
		return nil, err
	}
	return data, nil
}

type FuncParamData struct {
	RawParam   *parser.FuncParam
	Flag       *types.Flag
	IsOptional bool
	Imports    []string
}

func extractDataFromFuncParam(param *parser.FuncParam) (*FuncParamData, error) {
	if !isParamTypeKnown(param.Type.Base.CoreType) {
		return nil, fmt.Errorf("type %s is unknown", param.Type.Base.CoreType)
	}
	data := &FuncParamData{
		RawParam: param,
		Flag: &types.Flag{
			Name: param.Name,
			Type: param.Type.ToString(),
		},
		IsOptional: !isParameterRequired(param),
		Imports:    getRequiredImportsForParam(param),
	}
	return data, nil
}

func generateShortFlagsNames(flags []*types.Flag) error {
	nameFlagDict := make(map[string]*types.Flag)
	names := make([]string, len(flags))
	for i, f := range flags {
		nameFlagDict[f.Name] = f
		names[i] = f.Name
	}
	shortNames, err := trie.GetShortNames(names)
	if err != nil {
		return err
	}
	if len(shortNames) != len(flags) {
		return fmt.Errorf("unexpected number of short names generated: expected %d, got %d", len(shortNames), len(flags))
	}
	for name, shortName := range shortNames {
		f, ok := nameFlagDict[name]
		if !ok {
			return fmt.Errorf("generated a short name for an unexpected argument %s", name)
		}
		f.ShortName = shortName
	}
	return nil
}

func isParamTypeKnown(paramType string) bool {
	validTypes := map[string]struct{}{
		"string":     {},
		"int":        {},
		"int8":       {},
		"int16":      {},
		"int32":      {},
		"int64":      {},
		"uint":       {},
		"uint8":      {},
		"uint16":     {},
		"uint32":     {},
		"uint64":     {},
		"bool":       {},
		"float32":    {},
		"float64":    {},
		"error":      {},
		"complex64":  {},
		"complex128": {},
		"byte":       {},
		"rune":       {},
	}
	_, ok := validTypes[paramType]
	return ok
}

func createMain(mainPkgFunction *parser.PackageFunctions) (string, error) {
	outs := make([]string, 0, len(mainPkgFunction.Functions))
	// TODO: move importsMap and castFuncsMap to output
	importsMap := make(map[string]struct{})
	castFuncsMap := make(map[string]string)
	indirFuncsMap := make(map[string]string)
	predefinedFuncsMap := make(map[string]string)
	treatedFunctions := make([]*parser.PkgFunc, 0, len(mainPkgFunction.Functions))
	var shouldAppendParsingFunctions bool
	for _, rawFn := range mainPkgFunction.Functions {
		processedFn, err := extractDataFromParsedFunction(rawFn)
		if err != nil {
			log.Printf("[WARN]: skipping function %s: %v", rawFn.Name, err)
			continue
		}
		for imp := range processedFn.Imports {
			if _, ok := importsMap[imp]; !ok {
				importsMap[imp] = struct{}{}
			}
		}
		out, err := generateFromFunction(processedFn, castFuncsMap, indirFuncsMap, predefinedFuncsMap)
		if len(out) != 0 {
			shouldAppendParsingFunctions = true
		}
		if err != nil {
			log.Printf("[WARN]: skipping function %s: %v", rawFn.Name, err)
			continue
		}
		outs = append(outs, out)
		treatedFunctions = append(treatedFunctions, rawFn)
	}
	if shouldAppendParsingFunctions {
		outs = append(outs, gosifFuncs)
	}
	out := strings.Join(outs, "\n")
	mainOut, err := generateMainFunc(treatedFunctions, mainPkgFunction.HasMain)
	if err != nil {
		return "", err
	}
	imports := make([]string, 0, len(importsMap))
	for imp := range importsMap {
		imports = append(imports, imp)
	}
	fullOut, err := generateFromTemplate(tmplFullFile, &tmplFullFileInput{
		MainFunc:        mainOut,
		Imports:         imports,
		Out:             out,
		CastFuncs:       castFuncsMap,
		IndirFuncs:      indirFuncsMap,
		PredefinedFuncs: predefinedFuncsMap,
	})
	if err != nil {
		return "", nil
	}
	fullOutFormatted, err := formatOutput(fullOut)
	if err != nil {
		log.Printf("[WARN] failed to format the generated code: %v", err)
		return fullOut, nil
	}
	return fullOutFormatted, nil
}

func generateHelpFunctions(functions []*parser.PkgFunc) (string, error) {
	scriptsHelpFunc, err := generateScriptsHelpFunction(functions)
	return scriptsHelpFunc, err
}

func generateScriptsHelpFunction(functions []*parser.PkgFunc) (string, error) {
	scriptsNames := make([]string, len(functions))
	for i, f := range functions {
		scriptsNames[i] = f.Name
	}
	in := &tmplScriptsHelpFunctionInput{
		ScriptsNames: scriptsNames,
	}
	helpFunc, err := generateFromTemplate(tmplScriptsHelpFunction, in)
	return helpFunc, err
}

// TODO: Refactor
func generateFromFunction(fn *FuncForGenerator, castFuncsMap map[string]string, indirFuncsMap map[string]string, predefinedFuncsMap map[string]string) (string, error) {
	if len(fn.ParsedFunc.Parameters) == 0 {
		return "", nil
	}
	cases := make([]string, len(fn.OptionalParams)+len(fn.RequiredParams))
	params := fn.RequiredParams[:]
	params = append(params, fn.OptionalParams...)
	for i, param := range params {
		// TODO: generate during parameter parsing
		paramCase, err := generateCase(param.RawParam, param.Flag)
		if err != nil {
			return "", fmt.Errorf("case generation failed: %v", err)
		}
		cases[i] = paramCase
		if _, ok := castFuncsMap[param.RawParam.Type.Base.CoreType]; !ok {
			// TODO: do not pass castFuncsMap to this function
			castFn, err := generateCastFunction(param.RawParam.Type.Base.CoreType, castFuncsMap, predefinedFuncsMap)
			if err != nil {
				return "", fmt.Errorf("generating a cast function failed: %v", err)
			}
			if param.RawParam.Type.Base.CoreType != "byte" {
				castFuncsMap[param.RawParam.Type.Base.CoreType] = castFn
			}
		}
		if err := generateIndirFuncs(param.RawParam, indirFuncsMap); err != nil {
			return "", err
		}
	}
	if len(fn.RequiredParams) != 0 {
		predefinedFuncsMap[funcCheckRequiredFlags.name] = funcCheckRequiredFlags.body
	}
	flags, err := composeFlagsList(params, fn)
	if err != nil {
		return "", err
	}
	requiredFlags := make([]types.Flag, 0, len(fn.RequiredParams))
	for _, p := range fn.RequiredParams {
		requiredFlags = append(requiredFlags, *p.Flag)
	}
	flagStructTmplInput := &funcFlagStructureTmplInput{
		Flags:        flags,
		FunctionName: fn.ParsedFunc.Name,
	}
	out1, err := generateFromTemplate(tmplFuncFlagsStruct, flagStructTmplInput)
	if err != nil {
		return "", err
	}
	parseFlagsFuncTmplIn := &tmplParseFlagsFuncInput{
		Cases:         cases,
		FuncFlags:     flags,
		RequiredFlags: requiredFlags,
		FunctionName:  fn.ParsedFunc.Name,
	}
	out2, err := generateFromTemplate(tmplParseFlagsFunc, parseFlagsFuncTmplIn)
	if err != nil {
		return "", err
	}
	generateRunScriptFuncTmplInput := (*runScriptFuncTmplInput)(flagStructTmplInput)
	out3, err := generateFromTemplate(tmplRunScriptFunc, generateRunScriptFuncTmplInput)
	if err != nil {
		return "", err
	}
	funcHelp, err := generateFuncHelpFunction(fn.ParsedFunc, flags, requiredFlags)
	if err != nil {
		return "", err
	}
	out := strings.Join([]string{out1, out2, out3, funcHelp}, "\n")
	return out, nil
}

func composeFlagsList(params []*FuncParamData, fn *FuncForGenerator) ([]types.Flag, error) {
	flags := make([]types.Flag, 0, len(params))
	for _, p := range fn.ParsedFunc.Parameters {
		found := false
		for _, rf := range fn.RequiredParams {
			if rf.RawParam.Name == p.Name {
				flags = append(flags, *rf.Flag)
				found = true
				break
			}
		}
		if found {
			continue
		}
		for _, of := range fn.OptionalParams {
			if of.RawParam.Name == p.Name {
				flags = append(flags, *of.Flag)
				found = true
				break
			}
		}
		if found {
			continue
		}
		return nil, fmt.Errorf("flag %s not found", p.Name)
	}
	return flags, nil
}

func generateFuncHelpFunction(fn *parser.PkgFunc, flags []types.Flag, requiredFlags []types.Flag) (string, error) {
	in := &tmplFuncHelpFunctionInput{
		FunctionName:  fn.Name,
		Flags:         flagsToHelpFlags(flags),
		RequiredFlags: flagsToHelpFlags(requiredFlags),
	}
	out, err := generateFromTemplate(tmplFuncHelpFunction, in)
	if err != nil {
		return "", err
	}
	return out, nil
}

func flagsToHelpFlags(flags []types.Flag) []helpFlagData {
	helpFlags := make([]helpFlagData, len(flags))
	for i, f := range flags {
		helpFlags[i] = helpFlagData{
			Name: f.Name,
			Type: f.Type,
		}
		if f.ShortName != nil && *f.ShortName != f.Name {
			helpFlags[i].ShortName = f.ShortName
		}
	}
	return helpFlags
}

//TODO: refactor
func generateIndirFuncs(param *parser.FuncParam, indirFuncsMap map[string]string) error {
	indirArrFuncsNames := make([]string, 0)
	indirArrFuncInputs := make([]*tmplIndirArrFunctionInput, 0)
	for i, l := range param.Type.Layers {
		if l.ArrayConfig == nil {
			return fmt.Errorf("found layer with a nil array config")
		}
		if l.IndirectionLevel == 0 {
			continue
		}
		in := &tmplIndirArrFunctionNameInput{
			IndirectionLevel: l.IndirectionLevel,
			ArrInfo: tmplArrayInfo{
				IsSlice:     l.ArrayConfig.IsSlice,
				ArrayLength: l.ArrayConfig.Length,
				ElType:      param.Type.LayerElType(i),
			},
		}
		name, err := generateFromTemplate(tmplIndirArrFunctionName, in)
		if err != nil {
			return fmt.Errorf("failed to generate an indirection function name: %v", err)
		}
		indirArrFuncsNames = append(indirArrFuncsNames, name)
		indirFuncIn := (tmplIndirArrFunctionInput)(*in)
		indirArrFuncInputs = append(indirArrFuncInputs, &indirFuncIn)
	}

	for i, name := range indirArrFuncsNames {
		if _, ok := indirFuncsMap[name]; ok {
			continue
		}
		var err error
		indirFuncsMap[name], err = generateFromTemplate(tmplIndirArrFunction, indirArrFuncInputs[i])
		if err != nil {
			return fmt.Errorf("failed to generate an indirection function: %v", err)
		}
	}
	if param.Type.Base.IndirectionLevel == 0 {
		return nil
	}
	in := &tmplIndirFunctionNameInput{
		Type:             param.Type.Base.CoreType,
		IndirectionLevel: param.Type.Base.IndirectionLevel,
	}
	indirBaseFuncName, err := generateFromTemplate(tmplIndirFunctionName, in)
	if err != nil {
		return fmt.Errorf("failed to generate an indirection function name: %v", err)
	}
	indirBaseFuncIn := (tmplIndirFunctionInput)(*in)
	if _, ok := indirFuncsMap[indirBaseFuncName]; ok {
		return nil
	}
	indirFuncsMap[indirBaseFuncName], err = generateFromTemplate(tmplIndirFunction, &indirBaseFuncIn)
	if err != nil {
		return err
	}
	return nil
}

func isParameterRequired(p *parser.FuncParam) bool {
	if p.Type.Base.CoreType == "bool" && !p.IsAnArray() {
		return false
	}
	if p.Type.IsPointer {
		return false
	}
	return true
}

func getRequiredImportsForParam(param *parser.FuncParam) []string {
	imports := make([]string, 0, 1)
	switch param.Type.Base.CoreType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "byte":
		imports = append(imports, "strconv")
	case "bool":
		imports = append(imports, "strings")
	}
	return imports
}

func generateCase(param *parser.FuncParam, f *types.Flag) (string, error) {
	if len(param.Type.Layers) > 1 {
		return "", fmt.Errorf("NYI")
	}
	prefix, err := generateCasePrefix(param, f)
	if err != nil {
		return "", err
	}
	castType := param.Type.Base.CoreType
	if castType == "byte" {
		castType = "uint8"
	}
	tmplArgCastIn := &tmplArgCastInput{
		IndirectionLevel: param.Type.Base.IndirectionLevel,
		Type:             castType,
		InArray:          param.IsAnArray(),
	}
	argParsing, err := generateFromTemplate(tmplArgCast, &tmplArgCastIn)
	if err != nil {
		return "", err
	}
	if param.IsAnArray() {
		arrLayer := param.Type.Layers[0]
		tmplArrayLayerIn := &tmplArrayLayerInput{
			ArrInfo: tmplArrayInfo{
				IsSlice: arrLayer.ArrayConfig.IsSlice,
				ElType:  param.Type.Base.ToString(),
			},
			IndirectionLevel:     arrLayer.IndirectionLevel,
			Payload:              argParsing,
			BaseIndirectionLevel: param.Type.Base.IndirectionLevel,
		}
		if !arrLayer.ArrayConfig.IsSlice {
			tmplArrayLayerIn.ArrInfo.ArrayLength = arrLayer.ArrayConfig.Length
		}
		argParsing, err = generateFromTemplate(tmplArrayLayer, tmplArrayLayerIn)
		if err != nil {
			return "", err
		}
	}
	tmplArgCastPostfixIn := &tmplArgCastPostfixInput{
		FlagName:  f.Name,
		InArray:   param.IsAnArray(),
		IsPointer: param.Type.IsPointer,
		BaseType:  param.Type.Base.CoreType,
	}
	postfix, err := generateFromTemplate(tmplArgCastPostfix, tmplArgCastPostfixIn)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%s", prefix, argParsing, postfix), nil
}

func generateCasePrefix(param *parser.FuncParam, f *types.Flag) (string, error) {
	layersCount := len(param.Type.Layers)
	prefixIn := &tmplArgCastPrefixInput{
		FlagName:    f.Name,
		LayersCount: layersCount,
		BaseType:    param.Type.Base.CoreType,
	}
	prefix, err := generateFromTemplate(tmplArgCastPrefix, prefixIn)
	if err != nil {
		return "", err
	}
	return prefix, nil
}

func generateCastFunction(paramCoreType string, castFuncsMap map[string]string, predefinedFuncsMap map[string]string) (string, error) {
	baseIn := castFunctionBaseInput{
		Type: paramCoreType,
	}
	switch paramCoreType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		bitSize, err := getNumTypeBitSize(paramCoreType)
		if err != nil {
			return "", err
		}
		var signed bool
		if paramCoreType[0] == 'i' {
			signed = true
		}
		in := &castFuncNumInput{
			BitSize:               bitSize,
			Signed:                signed,
			castFunctionBaseInput: baseIn,
		}
		return generateFromTemplate(tmplCastFunctionNum, in)
	case "byte":
		if _, ok := castFuncsMap["uint8"]; ok {
			return "", nil
		}
		in := &castFuncNumInput{
			BitSize: "8",
			Signed:  false,
			castFunctionBaseInput: castFunctionBaseInput{
				Type: "uint8",
			},
		}
		uint8CastFunc, err := generateFromTemplate(tmplCastFunctionNum, in)
		if err != nil {
			return "", err
		}
		castFuncsMap["uint8"] = uint8CastFunc
		return "", nil
	case "rune":
		return generateFromTemplate(tmplCastFunctionRune, baseIn)
	case "float32", "float64":
		bitSize, err := getNumTypeBitSize(paramCoreType)
		if err != nil {
			return "", err
		}
		in := &castFuncFloatInput{
			BitSize:               bitSize,
			castFunctionBaseInput: baseIn,
		}
		return generateFromTemplate(tmplCastFunctionFloat, in)
	case "string":
		return generateFromTemplate(tmplCastFunctionString, baseIn)
	case "bool":
		return generateFromTemplate(tmplCastFunctionBool, baseIn)
	case "error":
		return generateFromTemplate(tmplCastFunctionError, baseIn)
	case "complex64", "complex128":
		if _, ok := predefinedFuncsMap[funcStringParseArgAsComplex.name]; !ok {
			predefinedFuncsMap[funcStringParseArgAsComplex.name] = funcStringParseArgAsComplex.body
		}
		auxCastType := "float32"
		if paramCoreType == "complex128" {
			auxCastType = "float64"
		}
		if _, ok := castFuncsMap[auxCastType]; !ok {
			var err error
			castFuncsMap[auxCastType], err = generateCastFunction(auxCastType, castFuncsMap, predefinedFuncsMap)
			if err != nil {
				return "", err
			}
		}
		bitSize, err := getNumTypeBitSize(paramCoreType)
		if err != nil {
			return "", err
		}
		in := castFuncComplexInput{
			BitSize:               bitSize,
			castFunctionBaseInput: baseIn,
		}
		return generateFromTemplate(tmplCastFunctionComplex, in)
	default:
		return "", fmt.Errorf("cannot generate a cast function for type %s: NYI", paramCoreType)
	}
}

func getNumTypeBitSize(coreType string) (string, error) {
	bitSize := "0"
	offset := 0
	unknownTypeErr := fmt.Errorf("an unknown numerical core type %s, generator input cannot be composed", coreType)
	if len(coreType) < 3 {
		return "", unknownTypeErr
	}
	switch coreType[0:3] {
	case "int":
		offset = 3
	case "uin":
		offset = 4
	case "flo":
		offset = 5
	case "com":
		offset = 7
	default:
		return "", unknownTypeErr
	}
	if len(coreType) > offset {
		bitSize = coreType[offset:]
	}
	return bitSize, nil
}

func formatOutput(out string) (string, error) {
	outFormatted, err := format.Source([]byte(out))
	if err != nil {
		return "", fmt.Errorf("failed to format output: %v", err)
	}
	return string(outFormatted), nil
}

func newFlag(p *parser.FuncParam) *types.Flag {
	f := &types.Flag{
		Name: p.Name,
		Type: p.Type.ToString(),
	}
	return f
}

func generateMainFunc(scriptFuncs []*parser.PkgFunc, hasMain bool) (string, error) {
	cases := make([]string, len(scriptFuncs))
	for i, fn := range scriptFuncs {
		var err error
		cases[i], err = generateMainFuncCase(fn)
		if err != nil {
			return "", err
		}
	}
	mainIn := &mainFuncTmplInput{
		Cases:   cases,
		HasMain: hasMain,
	}
	return generateFromTemplate(tmplMainFunc, mainIn)
}

func generateMainFuncCase(scriptFunc *parser.PkgFunc) (string, error) {
	in := &mainFuncScriptCaseTmplInput{
		FunctionName: scriptFunc.Name,
	}
	if len(scriptFunc.Parameters) == 0 {
		scriptCase, err := generateFromTemplate(tmplMainFuncNoArgsScriptCase, in)
		return scriptCase, err
	}
	scriptCase, err := generateFromTemplate(tmplMainFuncScriptCase, in)
	return scriptCase, err
}

func generateFromTemplate(tmpl *template.Template, in interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func removePreviousOutput(dir string) error {
	filePath := getFilePath(dir)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to lookup the file %s: %v", filePath, err)
	}
	funcs, err := parser.GetFileFunctions(filePath)
	if err != nil {
		return err
	}
	for _, f := range funcs {
		if f.Name != "main" && f.Name != "gosif" && (len(f.Name) < 6 || f.Name[:6] != "gosif_") {
			return fmt.Errorf("failed to remove the file %s: it contains functions that were not generated by gosif", filePath)
		}
	}
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove the file %s: %v", filePath, err)
	}
	return nil
}

func writeToFile(filePath string, str string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(str); err != nil {
		return err
	}
	return nil
}

func getFilePath(dir string) string {
	return path.Join(dir, outFileName)
}
