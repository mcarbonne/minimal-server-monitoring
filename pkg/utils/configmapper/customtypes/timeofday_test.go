package customtypes_test

import (
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeOfDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseTimeOfDay() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeOfDay_String(t *testing.T) {
	tod := customtypes.TimeOfDay{Hour: 14, Minute: 30}
	expected := "14:30"

	if tod.String() != expected {
		t.Errorf("String() = %s; want %s", tod.String(), expected)
	}
}
