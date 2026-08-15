package cmds

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDiffGolden is an integration test mirroring diff.tape: it diffs the three
// pokemon fixtures (_cynd/_quil/_typh) through the real xplr diff command and
// verifies the emitted JSON matches golden.json byte for byte. Any behavioral
// change to parsing, diffing, or rendering that alters the output will fail
// here, prompting a regenerated golden file in testdata/.
func TestDiffGolden(t *testing.T) {
	fixtures := filepath.Join("testdata", "diff")
	files := []string{
		filepath.Join(fixtures, "155.json"),
		filepath.Join(fixtures, "156.json"),
		filepath.Join(fixtures, "157.json"),
	}
	want, err := os.ReadFile(filepath.Join(fixtures, "golden.json"))
	require.NoError(t, err)

	cmd := NewDiffCmd()
	cmd.SetArgs([]string{
		"-o", "json",
		"-f", files[0],
		"-f", files[1],
		"-f", files[2],
		"-k", "_cynd", "-k", "_quil", "-k", "_typh",
	})

	got := captureStdout(func() {
		require.NoError(t, cmd.Execute())
	})
	require.Equal(t, string(want), got)
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns the
// captured bytes. It restores the original os.Stdout when fn returns.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	return buf.String()
}
