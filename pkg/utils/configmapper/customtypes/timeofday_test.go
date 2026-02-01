package customtypes_test

import (
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
	"gotest.tools/v3/assert"
)

func TestParseTimeOfDay(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    customtypes.TimeOfDay
		wantErr bool
	}{
		{
			name:    "Valid morning time",
			input:   "08:00",
			want:    customtypes.TimeOfDay{Hour: 8, Minute: 0},
			wantErr: false,
		},
		{
			name:    "Valid evening time",
			input:   "23:59",
			want:    customtypes.TimeOfDay{Hour: 23, Minute: 59},
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "8h00",
			want:    customtypes.TimeOfDay{},
			wantErr: true,
		},
		{
			name:    "Out of range hours",
			input:   "24:00",
			want:    customtypes.TimeOfDay{},
			wantErr: true,
		},
		{
			name:    "Out of range minutes",
			input:   "12:60",
			want:    customtypes.TimeOfDay{},
			wantErr: true,
		},
		{
			name:    "Out of range minutes",
			input:   "12:-5",
			want:    customtypes.TimeOfDay{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := customtypes.ParseTimeOfDay(tt.input)
			if tt.wantErr {
				assert.Assert(t, err != nil, "Expected error for input %v", tt.input)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestTimeOfDay_String(t *testing.T) {
	tod := customtypes.TimeOfDay{Hour: 14, Minute: 30}
	expected := "14:30"
	assert.Equal(t, tod.String(), expected)
}
