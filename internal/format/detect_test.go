package format

import (
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  FormatType
	}{
		// JSON cases
		{"json object", `{"value": 42}`, FormatJSON},
		{"json array", `[1, 2, 3]`, FormatJSON},
		{"json nested", `{"cpu": {"user": 10, "system": 5}}`, FormatJSON},

		// CSV cases
		{"csv labeled", "cpu,75.5", FormatCSV},
		{"csv multi-value", "1,2,3,4,5", FormatCSV},
		{"csv tab separated", "memory\t1024", FormatCSV},

		// Raw cases
		{"raw number", "42", FormatRaw},
		{"raw float", "3.14159", FormatRaw},
		{"raw negative", "-17.5", FormatRaw},
		{"raw with units", "45.2ms", FormatRaw},
		{"empty string", "", FormatRaw},
		{"whitespace", "   ", FormatRaw},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.input)
			if got != tt.want {
				t.Errorf("Detect(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAutoParse_JSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantValue float64
		wantLabel string
	}{
		{
			name:      "simple value field",
			input:     `{"value": 42.5}`,
			wantLen:   1,
			wantValue: 42.5,
		},
		{
			name:      "with label",
			input:     `{"metric": "cpu", "value": 75}`,
			wantLen:   1,
			wantValue: 75,
			wantLabel: "cpu",
		},
		{
			name:    "multiple numeric fields",
			input:   `{"cpu": 45, "memory": 1024}`,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AutoParse(tt.input)

			if result.Format != FormatJSON {
				t.Errorf("expected JSON format, got %v", result.Format)
			}

			if len(result.Points) != tt.wantLen {
				t.Errorf("expected %d points, got %d", tt.wantLen, len(result.Points))
				return
			}

			if tt.wantLen == 1 {
				if result.Points[0].Value != tt.wantValue {
					t.Errorf("expected value %.2f, got %.2f", tt.wantValue, result.Points[0].Value)
				}
				if tt.wantLabel != "" && result.Points[0].Label != tt.wantLabel {
					t.Errorf("expected label %q, got %q", tt.wantLabel, result.Points[0].Label)
				}
			}
		})
	}
}

func TestAutoParse_CSV(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLen    int
		wantValues []float64
		wantLabels []string
	}{
		{
			name:       "labeled value",
			input:      "cpu,75.5",
			wantLen:    1,
			wantValues: []float64{75.5},
			wantLabels: []string{"cpu"},
		},
		{
			name:       "multiple numbers",
			input:      "10,20,30",
			wantLen:    3,
			wantValues: []float64{10, 20, 30},
		},
		{
			name:       "tab separated",
			input:      "memory\t2048",
			wantLen:    1,
			wantValues: []float64{2048},
			wantLabels: []string{"memory"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AutoParse(tt.input)

			if result.Format != FormatCSV {
				t.Errorf("expected CSV format, got %v", result.Format)
			}

			if len(result.Points) != tt.wantLen {
				t.Errorf("expected %d points, got %d", tt.wantLen, len(result.Points))
				return
			}

			for i, p := range result.Points {
				if i < len(tt.wantValues) && p.Value != tt.wantValues[i] {
					t.Errorf("point %d: expected value %.2f, got %.2f", i, tt.wantValues[i], p.Value)
				}
				if tt.wantLabels != nil && i < len(tt.wantLabels) && p.Label != tt.wantLabels[i] {
					t.Errorf("point %d: expected label %q, got %q", i, tt.wantLabels[i], p.Label)
				}
			}
		})
	}
}

func TestAutoParse_Raw(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantValue float64
	}{
		{
			name:      "integer",
			input:     "42",
			wantLen:   1,
			wantValue: 42,
		},
		{
			name:      "float",
			input:     "3.14159",
			wantLen:   1,
			wantValue: 3.14159,
		},
		{
			name:      "negative",
			input:     "-17.5",
			wantLen:   1,
			wantValue: -17.5,
		},
		{
			name:      "with units",
			input:     "Response time: 45.2ms",
			wantLen:   1, // "45.2" extracted
			wantValue: 45.2,
		},
		{
			name:    "empty",
			input:   "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AutoParse(tt.input)

			if result.Format != FormatRaw {
				t.Errorf("expected Raw format, got %v", result.Format)
			}

			if len(result.Points) != tt.wantLen {
				t.Errorf("expected %d points, got %d", tt.wantLen, len(result.Points))
				return
			}

			if tt.wantLen == 1 && result.Points[0].Value != tt.wantValue {
				t.Errorf("expected value %.5f, got %.5f", tt.wantValue, result.Points[0].Value)
			}
		})
	}
}
