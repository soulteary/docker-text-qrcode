/**
 * Copyright 2025 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	main "github.com/soulteary/docker-text-qrcode"
)

func TestFindQRExecutable(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Find qr executable",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := main.FindQRExecutable() // 注意：需要导出这个函数
			if tt.wantErr && err == nil {
				t.Errorf("FindQRExecutable() expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("FindQRExecutable() error = %v", err)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("FindQRExecutable() got empty path")
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "Echo test",
			args:    []string{"-c", "echo 'hello'"},
			want:    "hello\n",
			wantErr: false,
		},
		{
			name:    "Invalid command",
			args:    []string{"-c", "invalidcommand"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := main.Execute(tt.args...) // 注意：需要导出这个函数
			if tt.wantErr && err == nil {
				t.Errorf("Execute() expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := main.SetupRouter() // 注意：需要新增并导出这个函数

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "pre style=\"line-height: 0.92;user-select: none;text-shadow: 0 0 black;\"") {
		t.Error("Response does not contain expected style")
	}
	if !strings.Contains(body, "style=\"filter: invert(1);background: #fff;\"") {
		t.Error("Response does not contain dark mode style")
	}
}

func TestTemplates(t *testing.T) {
	t.Run("Test default template format", func(t *testing.T) {
		output := "test content"
		result := fmt.Sprintf(main.DefaultTemplate, main.DarkTemplate, output) // 注意：需要导出这些常量

		expectedParts := []string{
			"<!DOCTYPE html>",
			"<html lang=\"en\"",
			output,
			main.DarkTemplate,
		}

		for _, part := range expectedParts {
			if !strings.Contains(result, part) {
				t.Errorf("Template result does not contain expected part: %s", part)
			}
		}
	})

	t.Run("Test dark template content", func(t *testing.T) {
		expected := ` style="filter: invert(1);background: #fff;"`
		if main.DarkTemplate != expected {
			t.Errorf("DarkTemplate = %v, want %v", main.DarkTemplate, expected)
		}
	})
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	code := m.Run()
	os.Exit(code)
}
