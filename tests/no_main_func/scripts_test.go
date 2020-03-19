//+build integration_tests

package no_main_func

import (
	"fmt"
	"path"
	"testing"

	"github.com/SergeyShpak/gosif/tests/utils"
)

const outBin = "test_bin"
const outDir = "test"

func TestNoMainFunc(t *testing.T) {
	if err := utils.Setup(outBin, outDir, true); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cleanup := func() {
		utils.RemoveArtifacts(outBin, outDir)
	}
	t.Cleanup(cleanup)
	t.Run("Running scripts with the gosif function", func(t *testing.T) {
		cases := []utils.TestCase{
			{
				ExpectedOut: "it works!",
			},
		}
		binPath := path.Join(outDir, outBin)
		for i, tc := range cases {
			i, tc := i, tc
			t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
				t.Parallel()
				out, err := utils.RunScript(binPath, tc.ScriptName, tc.Args)
				if err := utils.CheckRunScriptResult(&tc, out, err); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}
