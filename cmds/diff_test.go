package cmds

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherInputs(t *testing.T) {
	dir := t.TempDir()
	writeFile := func(name, content string) string {
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
		return p
	}
	fileA := writeFile("a.json", "AAA")
	fileB := writeFile("b.json", "BBB")
	stdinFile := writeFile("stdin.json", "SSS")

	// os.DevNull is a character device, so piped() is false and stdin is
	// ignored; a regular file is not, so it is read as a piped source.
	devNull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { devNull.Close() })

	tests := []struct {
		name     string
		args     []string
		files    []string
		useStdin bool
		want     [][]byte
	}{
		{
			name:  "files only in flag order",
			files: []string{fileA, fileB},
			want:  [][]byte{[]byte("AAA"), []byte("BBB")},
		},
		{
			name: "args only as inline data",
			args: []string{"one", "two"},
			want: [][]byte{[]byte("one"), []byte("two")},
		},
		{
			name:  "files come before args",
			files: []string{fileA},
			args:  []string{"arg"},
			want:  [][]byte{[]byte("AAA"), []byte("arg")},
		},
		{
			name:     "stdin comes last",
			files:    []string{fileA},
			args:     []string{"arg"},
			useStdin: true,
			want:     [][]byte{[]byte("AAA"), []byte("arg"), []byte("SSS")},
		},
		{
			name:     "stdin only",
			useStdin: true,
			want:     [][]byte{[]byte("SSS")},
		},
		{
			// the unset -f flag arrives as an empty string; it must be
			// skipped so a piped stdin still supplies the data.
			name:     "empty file flag is skipped for stdin",
			files:    []string{""},
			useStdin: true,
			want:     [][]byte{[]byte("SSS")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := devNull
			if tt.useStdin {
				f, err := os.Open(stdinFile)
				require.NoError(t, err)
				t.Cleanup(func() { f.Close() })
				stdin = f
			}
			got, err := gatherInputs(tt.args, tt.files, stdin)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGatherInputsEmptyStdinSkipped(t *testing.T) {
	// a piped but empty stdin adds no operand.
	empty := filepath.Join(t.TempDir(), "empty")
	require.NoError(t, os.WriteFile(empty, nil, 0o600))
	f, err := os.Open(empty)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })

	got, err := gatherInputs([]string{"x"}, nil, f)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{[]byte("x")}, got)
}

func TestGatherInputsFileError(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { devNull.Close() })

	_, err = gatherInputs(nil, []string{filepath.Join(t.TempDir(), "missing.json")}, devNull)
	require.Error(t, err)
}

func TestDefaultKeys(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []string
	}{
		{name: "zero", n: 0, want: []string{}},
		{name: "one", n: 1, want: []string{"_f1"}},
		{name: "three", n: 3, want: []string{"_f1", "_f2", "_f3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, defaultKeys(tt.n))
		})
	}
}
