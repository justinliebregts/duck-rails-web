package main

import (
	"gofiber/config"
	"path/filepath"
	"testing"
)

func Test_loadDotEnv(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.LoadDotEnv(filepath.Base("."))
		})
	}
}
