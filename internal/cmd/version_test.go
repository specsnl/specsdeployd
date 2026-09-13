package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/specsnl/specsdeployd/internal/cmd"
)

func run(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var out, errOut bytes.Buffer

	root := cmd.NewRootCmd(cmd.NewApp())
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)

	err = root.Execute()

	return out.String(), errOut.String(), err
}

func TestVersion(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"the subcommand names the binary", []string{"version"}, "specsdeployd version " + cmd.Version + "\n"},
		{"the root flag is bare", []string{"--version"}, cmd.Version + "\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := run(t, tc.args...)
			if err != nil {
				t.Fatalf("%v: %v", tc.args, err)
			}

			if stdout != tc.want {
				t.Errorf("stdout = %q, want %q", stdout, tc.want)
			}

			if stderr != "" {
				t.Errorf("version narrated on stderr: %q", stderr)
			}
		})
	}
}

func TestVersion_DefaultsToDev(t *testing.T) {
	if cmd.Version != "dev" {
		t.Errorf("Version = %q, want %q", cmd.Version, "dev")
	}
}

// The build files name the variable in a string no compiler checks, so a rename
// would ship every release as "dev" with nothing failing.
func TestVersion_BuildFilesInjectThisVariable(t *testing.T) {
	const (
		module = "github.com/specsnl/specsdeployd"
		want   = module + "/internal/cmd.Version"
	)

	for _, file := range []string{"../../.goreleaser.yml", "../../Dockerfile"} {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		// The Dockerfile builds the path from ${GO_MODULE}, so match the suffix too.
		if !strings.Contains(string(content), want) &&
			!strings.Contains(string(content), strings.TrimPrefix(want, module)) {
			t.Errorf("%s does not inject %s", file, want)
		}
	}
}

func TestVersion_RejectsArguments(t *testing.T) {
	if _, _, err := run(t, "version", "extra"); err == nil {
		t.Fatal("want an error for an unexpected argument, got none")
	}
}

func TestRoot_BareInvocationPrintsHelp(t *testing.T) {
	stdout, _, err := run(t)
	if err != nil {
		t.Fatalf("bare invocation: %v", err)
	}

	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q, want the help", stdout)
	}
}

func TestRoot_UnknownCommandStillFails(t *testing.T) {
	if _, _, err := run(t, "bogus"); err == nil {
		t.Fatal("want an error for an unknown command, got none")
	}
}
