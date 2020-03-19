package parser

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type parsedFile struct {
	Contents *bufio.Reader
	Package  *ast.Package
}

type Packages struct {
	MainPackage   *PackageFunctions
	OtherPacakges []*PackageFunctions
}

func ParsePackagesFunctions(dir string) (*Packages, error) {
	pkgs, err := parseFiles(dir)
	if err != nil {
		return nil, err
	}
	packages := &Packages{}
	for pkgDir, pkgsInDir := range pkgs {
		if len(pkgsInDir) > 1 {
			log.Printf("%s: multiple packages per directory is not currently supported", pkgDir)
			continue
		}
		if len(pkgsInDir) == 0 {
			continue
		}
		for pkgName, pkg := range pkgsInDir {
			if pkgName != "main" {
				log.Printf("[WARN] ignoring package %s, only main is parsed", pkgName)
				continue
			}
			pkgFuncs, err := getPackageFunctions(pkgDir, pkgName, pkg)
			if err != nil {
				return nil, err
			}
			packages.MainPackage = pkgFuncs
			return packages, nil
		}
	}
	return packages, nil
}

func GetFileFunctions(path string) ([]*PkgFunc, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing the file \"%s\" failed: %v", path, err)
	}
	fileName := filepath.Base(path)
	funcs, err := getFunctionsFromFile(fileName, f)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}
	return funcs, nil
}

func parseFiles(dirPath string) (map[string]map[string]*ast.Package, error) {
	result := make(map[string]map[string]*ast.Package)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkgs, err := parser.ParseDir(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		result[path] = pkgs
		return nil
	}
	if err := filepath.Walk(dirPath, walkFn); err != nil {
		return nil, err
	}
	return result, nil
}

type arrayConfig struct {
	IsSlice bool
	Length  int
}

type parameterTypeLayer struct {
	ArrayConfig      *arrayConfig
	IndirectionLevel int
}

func (p *parameterTypeLayer) ToString() string {
	stars := strings.Repeat("*", p.IndirectionLevel)
	if p.ArrayConfig == nil {
		return stars
	}
	var slice string
	if p.ArrayConfig.IsSlice {
		slice = "[]"
	} else {
		slice = fmt.Sprintf("[%d]", p.ArrayConfig.Length)
	}
	return fmt.Sprintf("%s%s", stars, slice)
}

type parameterTypeBase struct {
	IndirectionLevel int
	CoreType         string
}

func (b *parameterTypeBase) ToString() string {
	return fmt.Sprintf("%s%s", strings.Repeat("*", b.IndirectionLevel), b.CoreType)
}

type parameterType struct {
	IsPointer bool
	Layers    []*parameterTypeLayer
	Base      parameterTypeBase
}

func (p *parameterType) ToString() string {
	var sb strings.Builder
	if p.IsPointer {
		sb.WriteRune('*')
	}
	for _, l := range p.Layers {
		sb.WriteString(l.ToString())
	}
	baseLevel := fmt.Sprintf("%s%s", strings.Repeat("*", p.Base.IndirectionLevel), p.Base.CoreType)
	sb.WriteString(baseLevel)
	return sb.String()
}

func (p *parameterType) LayerElType(layerInd int) string {
	if layerInd >= len(p.Layers) {
		return ""
	}
	if layerInd == len(p.Layers)-1 {
		return p.Base.ToString()
	}
	var sb strings.Builder
	for i := layerInd; i < len(p.Layers); i++ {
		sb.WriteString(p.Layers[i].ToString())
	}
	sb.WriteString(p.Base.ToString())
	return sb.String()
}

// TODO: refactor this function
func extractParameterType(astType ast.Expr) (*parameterType, error) {
	curExpr := astType
	var pt parameterType
	if starExpr, ok := astType.(*ast.StarExpr); ok {
		pt.IsPointer = true
		curExpr = starExpr.X
	}
	layers := make([]*parameterTypeLayer, 0)
	var curLayer *parameterTypeLayer
	for {
		switch t := curExpr.(type) {
		case *ast.Ident:
			base := parameterTypeBase{
				CoreType: t.Name,
			}
			if curLayer != nil {
				if curLayer.ArrayConfig == nil {
					base.IndirectionLevel = curLayer.IndirectionLevel
				} else {
					layers = append(layers, curLayer)
				}
			}
			pt.Layers = layers
			pt.Base = base
			return &pt, nil
		case *ast.StarExpr:
			if curLayer == nil {
				curLayer = &parameterTypeLayer{}
			}
			curLayer.IndirectionLevel++
			curExpr = t.X
		case *ast.ArrayType:
			if curLayer == nil {
				curLayer = &parameterTypeLayer{}
			}
			curLayer.ArrayConfig = &arrayConfig{
				IsSlice: true,
			}
			if t.Len != nil {
				switch tLen := t.Len.(type) {
				case *ast.BasicLit:
					val, err := strconv.Atoi(tLen.Value)
					if err != nil {
						return nil, fmt.Errorf("failed to cast array value %s to int: %v", tLen.Value, err)
					}
					curLayer.ArrayConfig.IsSlice = false
					curLayer.ArrayConfig.Length = val
				default:
					return nil, fmt.Errorf("parameter's array length expression %v is of an unknown type", t.Len)
				}
			}
			layers = append(layers, curLayer)
			curLayer = nil
			curExpr = t.Elt
		default:
			// TODO: return a better error
			return nil, fmt.Errorf("expected an array, a pointer or a basic type, got: %v", astType)
		}
	}
}

type PkgFunc struct {
	Name       string
	Parameters []*FuncParam
	IsExported bool
	Path       string
}

type PackageFunctions struct {
	PackageDir  string
	PackageName string
	Functions   []*PkgFunc
	HasMain     bool
}

func getPackageFunctions(pkgDir string, pkgName string, pkg *ast.Package) (*PackageFunctions, error) {
	packageFunctions := &PackageFunctions{
		PackageDir:  pkgDir,
		PackageName: pkgName,
	}
	funcs := make([]*PkgFunc, 0)
	for fileName, f := range pkg.Files {
		fileFuncs, err := getFunctionsFromFile(fileName, f)
		if err != nil {
			return nil, fmt.Errorf("failed to get functions from file %s: %v", fileName, err)
		}
		funcs = append(funcs, fileFuncs...)
	}
	filteredFuncs := make([]*PkgFunc, 0, len(funcs))
	for _, f := range funcs {
		if f.Name == "main" {
			packageFunctions.HasMain = true
		}
		if !f.IsExported {
			log.Printf("[WARN]: skipping an unexported function %s in the file %s", f.Name, f.Path)
			continue
		}
		filteredFuncs = append(filteredFuncs, f)
	}
	packageFunctions.Functions = filteredFuncs
	return packageFunctions, nil
}

func getFunctionsFromFile(fileName string, f *ast.File) ([]*PkgFunc, error) {
	if f == nil {
		return nil, fmt.Errorf("the passed *ast.File is nil")
	}
	funcs := make([]*PkgFunc, 0)
	for _, decl := range f.Decls {
		switch funcDecl := decl.(type) {
		case *ast.FuncDecl:
			pkgFunc := &PkgFunc{
				Name:       funcDecl.Name.Name,
				IsExported: funcDecl.Name.IsExported(),
				Path:       fileName,
			}
			parameters, err := parseFunction(funcDecl)
			if err != nil {
				log.Printf("[WARN]: skipping the function %s in %s: %v", funcDecl.Name.Name, fileName, err)
				continue
			}
			pkgFunc.Parameters = parameters
			funcs = append(funcs, pkgFunc)
		default:
			continue
		}
	}
	return funcs, nil
}

type FuncParam struct {
	Name string
	Type *parameterType
}

func (p FuncParam) IsAnArray() bool {
	return len(p.Type.Layers) > 0
}

func parseFunction(decl *ast.FuncDecl) ([]*FuncParam, error) {
	astParams := decl.Type.Params.List
	parameters := make([]*FuncParam, len(astParams))
	for i, param := range astParams {
		if len(param.Names) != 1 {
			return nil, fmt.Errorf("cannot parse a parameter with %d names", len(param.Names))
		}
		paramType, err := extractParameterType(param.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the parameter \"%s\": %v", param.Names[0].Name, err)
		}
		if err := checkParamType(paramType); err != nil {
			return nil, fmt.Errorf("failed to parse the parameter \"%s\": %v", param.Names[0].Name, err)
		}
		parameters[i] = &FuncParam{
			Name: param.Names[0].Name,
			Type: paramType,
		}
	}
	return parameters, nil
}

func checkParamType(p *parameterType) error {
	if len(p.Layers) > 1 {
		return fmt.Errorf("multidimensional parameters are not yet supported")
	}
	return nil
}

type parsedFn struct {
	FnName  string
	Package string
}
