package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/SergeyShpak/gosif/generator"
)

type TestCase struct {
	ScriptName  string
	Args        []string
	ExpectedOut string
	ExpectedErr error
}

func RemoveArtifacts(outBin string, outDir string) {
	toRemove := []string{"main.gen.go", outBin}
	for _, f := range toRemove {
		fileWithDir := path.Join(outDir, f)
		if err := os.RemoveAll(fileWithDir); err != nil {
			log.Printf("[WARN]: failed to remove the artifact %s", fileWithDir)
			return
		}
	}
}

func Setup(outBin string, outDir string, shouldRemoveArtifacts bool) error {
	if shouldRemoveArtifacts {
		RemoveArtifacts(outBin, outDir)
	}
	if err := generator.GenerateScriptsForDir(outDir); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", outBin)
	cmd.Dir = outDir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func RunScript(pathToBin string, scriptName string, args []string) (string, error) {
	var cmd *exec.Cmd
	if len(scriptName) != 0 {
		cmd = exec.Command(pathToBin, append([]string{scriptName}, args...)...)
	} else {
		cmd = exec.Command(pathToBin, args...)
	}
	var stdoutBuffer, stderrBuf bytes.Buffer
	cmd.Stdout = io.Writer(&stdoutBuffer)
	cmd.Stderr = io.Writer(&stderrBuf)
	cmd.Run()
	cmdOutStr := stdoutBuffer.String()
	var errToReturn error
	cmdErrStr := stderrBuf.String()
	if len(cmdErrStr) != 0 {
		errToReturn = fmt.Errorf(cmdErrStr)
	}
	return cmdOutStr, errToReturn
}

func CheckRunScriptResult(tc *TestCase, out string, err error) error {
	if err := checkRunScriptErr(tc.ExpectedErr, err); err != nil {
		return err
	}
	if out != tc.ExpectedOut {
		return fmt.Errorf("expected output \"%v\", but got \"%v\"", tc.ExpectedOut, out)
	}
	return nil
}

func CheckTestErrors(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if expectedErr == nil || actualErr == nil {
		return fmt.Errorf("expected err \"%v\", but got \"%v\"", expectedErr, actualErr)
	}
	if expectedErr.Error() != actualErr.Error() {
		return fmt.Errorf("expected an error \"%s\", got \"%s\"", expectedErr.Error(), actualErr.Error())
	}
	return nil
}

func checkRunScriptErr(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if expectedErr == nil || actualErr == nil {
		return fmt.Errorf("expected err \"%v\", but got \"%v\"", expectedErr, actualErr)
	}
	actualErrMsg := strings.Split(actualErr.Error(), "\n")[0]
	if expectedErr.Error() != actualErrMsg {
		return fmt.Errorf("expected err \"%v\", but got \"%s\"", expectedErr, actualErrMsg)
	}
	return nil
}
