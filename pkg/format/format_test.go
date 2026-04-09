package format

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAs(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		f       FormatType
		want    string
		wantErr bool
	}{
		{
			name: "json format",
			m:    map[string]any{"key": "value"},
			f:    FormatJson,
			want: `{"key":"value"}`,
		},
		{
			name: "yaml format",
			m:    map[string]any{"key": "value"},
			f:    FormatYaml,
			want: "key: value\n",
		},
		{
			name: "toml format",
			m:    map[string]any{"key": "value"},
			f:    FormatToml,
			want: "key = \"value\"\n",
		},
		{
			name:    "unsupported format returns error",
			m:       map[string]any{"key": "value"},
			f:       FormatType(99),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := As(tt.m, tt.f)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, string(got))
		})
	}
}
