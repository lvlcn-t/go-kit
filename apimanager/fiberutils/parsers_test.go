package fiberutils

import (
	"net"
	"testing"
	"time"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"Valid int", "123", int(123), false},
		{"Invalid int", "abc", int(0), true},
		{"Valid int8", "123", int8(123), false},
		{"Valid int16", "12345", int16(12345), false},
		{"Valid int32", "1234567890", int32(1234567890), false},
		{"Valid int64", "1234567890123456789", int64(1234567890123456789), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			var err error

			switch tt.want.(type) {
			case int:
				result, err = ParseInt[int](tt.input)
			case int8:
				result, err = ParseInt[int8](tt.input)
			case int16:
				result, err = ParseInt[int16](tt.input)
			case int32:
				result, err = ParseInt[int32](tt.input)
			case int64:
				result, err = ParseInt[int64](tt.input)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("ParseInt() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParseUint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"Valid uint", "123", uint(123), false},
		{"Invalid uint", "abc", uint(0), true},
		{"Valid uint8", "123", uint8(123), false},
		{"Valid uint16", "12345", uint16(12345), false},
		{"Valid uint32", "1234567890", uint32(1234567890), false},
		{"Valid uint64", "1234567890123456789", uint64(1234567890123456789), false},
		{"Valid uintptr", "123", uintptr(123), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			var err error

			switch tt.want.(type) {
			case uint:
				result, err = ParseUint[uint](tt.input)
			case uint8:
				result, err = ParseUint[uint8](tt.input)
			case uint16:
				result, err = ParseUint[uint16](tt.input)
			case uint32:
				result, err = ParseUint[uint32](tt.input)
			case uint64:
				result, err = ParseUint[uint64](tt.input)
			case uintptr:
				result, err = ParseUint[uintptr](tt.input)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("ParseUint() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParse_Float_Complex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
		funName string
	}{
		{"Valid float32", "123.45", float32(123.45), false, "ParseFloat"},
		{"Invalid float32", "abc", float32(0), true, "ParseFloat"},
		{"Valid float64", "123.45", float64(123.45), false, "ParseFloat"},
		{"Invalid float64", "abc", float64(0), true, "ParseFloat"},
		{"Valid complex64", "123.45+67.89i", complex64(123.45 + 67.89i), false, "ParseComplex"},
		{"Invalid complex64", "abc", complex64(0), true, "ParseComplex"},
		{"Valid complex128", "123.45+67.89i", complex128(123.45 + 67.89i), false, "ParseComplex"},
		{"Invalid complex128", "abc", complex128(0), true, "ParseComplex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			var err error

			switch tt.want.(type) {
			case float32:
				result, err = ParseFloat[float32](tt.input)
			case float64:
				result, err = ParseFloat[float64](tt.input)
			case complex64:
				result, err = ParseComplex[complex64](tt.input)
			case complex128:
				result, err = ParseComplex[complex128](tt.input)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("%s() error = %v, wantErr %v", tt.funName, err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("%s() = %v, want %v", tt.funName, result, tt.want)
			}
		})
	}
}

func TestParseIP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    net.IP
		wantErr bool
	}{
		{"Valid ipv4 address", "8.8.8.8", net.ParseIP("8.8.8.8"), false},
		{"Invalid complex64", "abc", nil, true},
		{"Valid ipv6 address", "::1", net.ParseIP("::1"), false},
		{"Invalid ipv6 address", "xyz", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIP(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !result.Equal(tt.want) {
				t.Errorf("ParseIP() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) { //nolint:dupl // Different types are tested.
	tests := []struct {
		name    string
		input   string
		formats []string
		want    time.Time
		wantErr bool
	}{
		{"Valid date default format", "2023-06-27", nil, time.Date(2023, 6, 27, 0, 0, 0, 0, time.UTC), false},
		{"Valid date custom format", "06/27/2023", []string{"01/02/2006"}, time.Date(2023, 6, 27, 0, 0, 0, 0, time.UTC), false},
		{"Invalid date", "invalid", nil, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := ParseDate(tt.formats...)
			result, err := parser(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.want) {
				t.Errorf("ParseDate() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		formats []string
		want    time.Time
		wantErr bool
	}{
		{"Valid time default format", "15:04:05", nil, parseTime(t, "15:04:05", time.TimeOnly), false},
		{"Valid time custom format", "03:04:05 PM", []string{"03:04:05 PM"}, parseTime(t, "03:04:05 PM", "03:04:05 PM"), false},
		{"Invalid time", "invalid", nil, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := ParseTime(tt.formats...)
			result, err := parser(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.want) {
				t.Errorf("ParseTime() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) { //nolint:dupl // Different types are tested.
	tests := []struct {
		name    string
		input   string
		formats []string
		want    time.Time
		wantErr bool
	}{
		{"Valid datetime default format", "2023-06-27T15:04:05Z", nil, time.Date(2023, 6, 27, 15, 4, 5, 0, time.UTC), false},
		{"Valid datetime custom format", "06/27/2023 03:04:05 PM", []string{"01/02/2006 03:04:05 PM"}, time.Date(2023, 6, 27, 15, 4, 5, 0, time.UTC), false},
		{"Invalid datetime", "invalid", nil, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := ParseDateTime(tt.formats...)
			result, err := parser(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.want) {
				t.Errorf("ParseDateTime() = %v, want %v", result, tt.want)
			}
		})
	}
}

// parseTime is a helper function to parse time with a format.
func parseTime(t *testing.T, value, format string) time.Time {
	t.Helper()
	v, err := time.Parse(format, value)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	return v
}
