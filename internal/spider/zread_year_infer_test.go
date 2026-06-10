package spider

import (
	"testing"
	"time"
)

func TestInferYear(t *testing.T) {
	tests := []struct {
		name     string
		mmdd     string
		now      time.Time
		expected int
		wantErr  bool
	}{
		{
			name:     "Same year, earlier month",
			mmdd:     "05/01",
			now:      time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			expected: 2026,
			wantErr:  false,
		},
		{
			name:     "Cross year (Dec to Jan)",
			mmdd:     "12/28",
			now:      time.Date(2027, 1, 3, 0, 0, 0, 0, time.UTC),
			expected: 2026,
			wantErr:  false,
		},
		{
			name:     "Same month",
			mmdd:     "06/01",
			now:      time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			expected: 2026,
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			mmdd:     "1",
			now:      time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Non-numeric month",
			mmdd:     "XX/01",
			now:      time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InferYear(tt.mmdd, tt.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("InferYear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("InferYear() got = %v, want %v", got, tt.expected)
			}
		})
	}
}
