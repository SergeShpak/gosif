package generator

import (
	"fmt"
)

type gosif_ReadFlag struct {
	PassedFlag string
	Args       []string
}

func gosif_ReadArgs(args []string, funcFlags map[string]struct{}) (map[string]gosif_ReadFlag, error) {
	if len(funcFlags) == 0 {
		return nil, fmt.Errorf("internal error: function flags map cannot be empty")
	}
	parsedFlags := make(map[string]gosif_ReadFlag)
	curPos := 0
	for curPos < len(args) {
		f := args[curPos]
		curPos++
		extractedFlag, err := gosif_UtilExtractFlag(f)
		if err != nil {
			return nil, fmt.Errorf("an error occurred during the flag \"%s\" extraction: %v", f, err)
		}
		if extractedFlag == "--" {
			return nil, fmt.Errorf("\"%s\" found, but ending command options and passing positional arguments is NYI", extractedFlag)
		}
		if _, ok := funcFlags[extractedFlag]; !ok {
			return nil, fmt.Errorf("an unexpected flag \"%s\" found", f)
		}
		flagArgs, err := gosif_UtilReadFlagArgs(args[curPos:], funcFlags)
		if err != nil {
			return nil, fmt.Errorf("an error occurred while parsing the flag %s arguments: %v", f, err)
		}
		curPos += len(flagArgs)
		parsedFlags[extractedFlag] = gosif_ReadFlag{
			PassedFlag: f,
			Args:       flagArgs,
		}
	}
	return parsedFlags, nil
}

func gosif_UtilExtractFlag(f string) (string, error) {
	if len(f) == 0 {
		return "", fmt.Errorf("internal error: expected a flag, got an empty string")
	}
	if f[0] != '-' {
		return "", fmt.Errorf("expected a flag (e.g. --flag), got an argument \"%s\"", f)
	}
	extracted := gosif_UtilExtractFlagAfterDash(f)
	if len(extracted) == 0 {
		return "", fmt.Errorf("passed flag \"%s\" is treated as empty and empty flags are not allowed", f)
	}
	return extracted, nil
}

func gosif_UtilExtractFlagAfterDash(f string) string {
	if f == "-" {
		return ""
	}
	if f == "--" {
		return "--"
	}
	if f[:2] != "--" {
		return f[1:]
	}
	return f[2:]
}

func gosif_UtilReadFlagArgs(args []string, funcFlags map[string]struct{}) ([]string, error) {
	flagArgs := make([]string, 0)
	for _, a := range args {
		if a[0] == '-' {
			extracted := gosif_UtilExtractFlagAfterDash(a)
			if extracted == "--" {
				break
			}
			if _, ok := funcFlags[extracted]; ok {
				break
			}
		}
		extractedArg := gosif_UtilExtractArg(a)
		if len(extractedArg) == 0 {
			flagArgs = append(flagArgs, extractedArg)
			continue
		}
		flagArgs = append(flagArgs, extractedArg)
	}
	return flagArgs, nil
}

func gosif_UtilExtractArg(arg string) string {
	if len(arg) < 2 {
		return arg
	}
	// arg is a quoted string, e.g. "-notAFlag" => -notAFlag
	if arg[0] == '"' && arg[len(arg)-1:] == "\"" {
		if len(arg) == 2 {
			return ""
		}
		return arg[1 : len(arg)-1]
	}
	if len(arg) < 4 {
		return arg
	}
	// args is a quoted string and quotes are escaped, e.g. \"quoted\" => "quoted"
	if arg[0:2] == "\\\"" && arg[len(arg)-2:] == "\\\"" {
		return arg[1:len(arg)-2] + "\""
	}
	return arg
}

// TODO: we need this function only if there are flags with a single argument
// We should filter it out in other cases
func gosif_GetFlagArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no arguments passed")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("expected a single argument, got %d (%v)", len(args), args)
	}
	return args[0], nil
}

// TODO: we need this function only if there are flags with a single boolean argument
// We should filter it out in other cases
func gosif_GetBoolArg(args []string) (string, error) {
	if len(args) == 0 {
		return "true", nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("expected a single argument, got %d (%v)", len(args), args)
	}
	return args[0], nil
}

const gosifFuncs = `
type gosif_ReadFlag struct {
	PassedFlag string
	Args       []string
}

func gosif_ReadArgs(args []string, funcFlags map[string]struct{}) (map[string]gosif_ReadFlag, error) {
	if len(funcFlags) == 0 {
		return nil, fmt.Errorf("internal error: function flags map cannot be empty")
	}
	parsedFlags := make(map[string]gosif_ReadFlag)
	curPos := 0
	for curPos < len(args) {
		f := args[curPos]
		curPos++
		extractedFlag, err := gosif_UtilExtractFlag(f)
		if err != nil {
			return nil, fmt.Errorf("an error occurred during the flag \"%s\" extraction: %v", f, err)
		}
		if extractedFlag == "--" {
			return nil, fmt.Errorf("\"%s\" found, but ending command options and passing positional arguments is NYI", extractedFlag)
		}
		if _, ok := funcFlags[extractedFlag]; !ok {
			return nil, fmt.Errorf("an unexpected flag \"%s\" found", f)
		}
		flagArgs, err := gosif_UtilReadFlagArgs(args[curPos:], funcFlags)
		if err != nil {
			return nil, fmt.Errorf("an error occurred while parsing the flag %s arguments: %v", f, err)
		}
		curPos += len(flagArgs)
		parsedFlags[extractedFlag] = gosif_ReadFlag{
			PassedFlag: f,
			Args:       flagArgs,
		}
	}
	return parsedFlags, nil
}

func gosif_UtilExtractFlag(f string) (string, error) {
	if len(f) == 0 {
		return "", fmt.Errorf("internal error: expected a flag, got an empty string")
	}
	if f[0] != '-' {
		return "", fmt.Errorf("expected a flag (e.g. --flag), got an argument \"%s\"", f)
	}
	extracted := gosif_UtilExtractFlagAfterDash(f)
	if len(extracted) == 0 {
		return "", fmt.Errorf("passed flag \"%s\" is treated as empty and empty flags are not allowed", f)
	}
	return extracted, nil
}

func gosif_UtilExtractFlagAfterDash(f string) string {
	if f == "-" {
		return ""
	}
	if f == "--" {
		return "--"
	}
	if f[:2] != "--" {
		return f[1:]
	}
	return f[2:]
}

func gosif_UtilReadFlagArgs(args []string, funcFlags map[string]struct{}) ([]string, error) {
	flagArgs := make([]string, 0)
	for _, a := range args {
		if a[0] == '-' {
			extracted := gosif_UtilExtractFlagAfterDash(a)
			if extracted == "--" {
				break
			}
			if _, ok := funcFlags[extracted]; ok {
				break
			}
		}
		extractedArg := gosif_UtilExtractArg(a)
		if len(extractedArg) == 0 {
			flagArgs = append(flagArgs, extractedArg)
			continue
		}
		flagArgs = append(flagArgs, extractedArg)
	}
	return flagArgs, nil
}

func gosif_UtilExtractArg(arg string) string {
	if len(arg) < 2 {
		return arg
	}
	// arg is a quoted string, e.g. "-notAFlag" => -notAFlag
	if arg[0] == '"' && arg[len(arg)-1:] == "\"" {
		if len(arg) == 2 {
			return ""
		}
		return arg[1 : len(arg)-1]
	}
	if len(arg) < 4 {
		return arg
	}
	// args is a quoted string and quotes are escaped, e.g. \"quoted\" => "quoted"
	if arg[0:2] == "\\\"" && arg[len(arg)-2:] == "\\\"" {
		return arg[1:len(arg)-2] + "\""
	}
	return arg
}

func gosif_GetFlagArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no arguments passed")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("expected a single argument, got %d (%v)", len(args), args)
	}
	return args[0], nil
}

func gosif_GetBoolArg(args []string) (string, error) {
	if len(args) == 0 {
		return "true", nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("expected a single argument, got %d (%v)", len(args), args)
	}
	return args[0], nil
}`
