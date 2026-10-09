package main

import (
	"os"
	"testing"
)

func TestResolvePort_Valid(t *testing.T) {
	tests := []struct {
		env  string
		want int
	}{
		{"", 8787},
		{"9000", 9000},
		{"8080", 8080},
		{"1", 1},
		{"65535", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			if tt.env != "" {
				os.Setenv("ASTROCYTE_PORT", tt.env)
				defer os.Unsetenv("ASTROCYTE_PORT")
			} else {
				os.Unsetenv("ASTROCYTE_PORT")
			}

			port, err := resolvePort()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if port != tt.want {
				t.Errorf("got %d, want %d", port, tt.want)
			}
		})
	}
}

func TestResolvePort_Invalid(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{"negative", "-1"},
		{"zero", "0"},
		{"too_high", "65536"},
		{"not_a_number", "abc"},
		{"float", "8080.5"},
		{"empty_string_with_space", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ASTROCYTE_PORT", tt.env)
			defer os.Unsetenv("ASTROCYTE_PORT")

			_, err := resolvePort()
			if err == nil {
				t.Fatalf("expected error for port %q, got nil", tt.env)
			}
		})
	}
}
