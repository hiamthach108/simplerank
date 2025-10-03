package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAppConfig(t *testing.T) {
	// Create a temporary test .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	envContent := `APP_NAME=test-app
APP_VERSION=1.0.0
HTTP_HOST=localhost
HTTP_PORT=8080
CACHE_DEFAULT_EXPIRE_TIME_SEC=3600
CACHE_CLEANUP_INTERVAL_HOUR=24
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=secret123
REDIS_DB=0`

	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	tests := []struct {
		name    string
		envDir  string
		wantErr bool
		verify  func(*testing.T, *AppConfig)
	}{
		{
			name:    "valid env file",
			envDir:  envFile,
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
				if config.App.Name != "test-app" {
					t.Errorf("Expected App.Name = test-app, got %s", config.App.Name)
				}
				if config.App.Version != "1.0.0" {
					t.Errorf("Expected App.Version = 1.0.0, got %s", config.App.Version)
				}
				if config.Server.Host != "localhost" {
					t.Errorf("Expected Server.Host = localhost, got %s", config.Server.Host)
				}
				if config.Server.Port != "8080" {
					t.Errorf("Expected Server.Port = 8080, got %s", config.Server.Port)
				}
				if config.Cache.DefaultExpireTimeSec != 3600 {
					t.Errorf("Expected Cache.DefaultExpireTimeSec = 3600, got %d", config.Cache.DefaultExpireTimeSec)
				}
				if config.Cache.CleanupIntervalHour != 24 {
					t.Errorf("Expected Cache.CleanupIntervalHour = 24, got %d", config.Cache.CleanupIntervalHour)
				}
				if config.Cache.RedisHost != "localhost" {
					t.Errorf("Expected Cache.RedisHost = localhost, got %s", config.Cache.RedisHost)
				}
				if config.Cache.RedisPort != "6379" {
					t.Errorf("Expected Cache.RedisPort = 6379, got %s", config.Cache.RedisPort)
				}
				if config.Cache.RedisPassword != "secret123" {
					t.Errorf("Expected Cache.RedisPassword = secret123, got %s", config.Cache.RedisPassword)
				}
				if config.Cache.RedisDB != 0 {
					t.Errorf("Expected Cache.RedisDB = 0, got %d", config.Cache.RedisDB)
				}
			},
		},
		{
			name:    "non-existent file",
			envDir:  "/non/existent/file.env",
			wantErr: true,
			verify:  nil,
		},
		{
			name:    "empty file path",
			envDir:  "",
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewAppConfig(tt.envDir)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("Expected config but got nil")
				return
			}

			if tt.verify != nil {
				tt.verify(t, config)
			}
		})
	}
}

func TestAppConfigPartialEnv(t *testing.T) {
	// Test with partial environment variables
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	// Only set some environment variables
	envContent := `APP_NAME=partial-app
HTTP_PORT=9000
REDIS_HOST=redis-server`

	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	config, err := NewAppConfig(envFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify set values
	if config.App.Name != "partial-app" {
		t.Errorf("Expected App.Name = partial-app, got %s", config.App.Name)
	}

	if config.Server.Port != "9000" {
		t.Errorf("Expected Server.Port = 9000, got %s", config.Server.Port)
	}

	if config.Cache.RedisHost != "redis-server" {
		t.Errorf("Expected Cache.RedisHost = redis-server, got %s", config.Cache.RedisHost)
	}

	// Verify unset values are zero values
	if config.App.Version != "" {
		t.Errorf("Expected App.Version to be empty, got %s", config.App.Version)
	}

	if config.Server.Host != "" {
		t.Errorf("Expected Server.Host to be empty, got %s", config.Server.Host)
	}

	if config.Cache.DefaultExpireTimeSec != 0 {
		t.Errorf("Expected Cache.DefaultExpireTimeSec to be 0, got %d", config.Cache.DefaultExpireTimeSec)
	}
}

func TestAppConfigInvalidEnvFile(t *testing.T) {
	// Test with malformed env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	// Invalid format - missing values, invalid syntax
	envContent := `APP_NAME=
HTTP_PORT=invalid_port_number
CACHE_DEFAULT_EXPIRE_TIME_SEC=not_a_number`

	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	config, err := NewAppConfig(envFile)
	if err == nil {
		t.Error("Expected error due to invalid env content, but got none")
	}

	// The function should return a config even if there are parsing errors
	if config == nil {
		t.Error("Expected config but got nil")
		return
	}

	// Check that string values are handled correctly
	if config.App.Name != "" {
		t.Errorf("Expected App.Name to be empty, got %s", config.App.Name)
	}

	// Port should be the invalid string value
	if config.Server.Port != "invalid_port_number" {
		t.Errorf("Expected Server.Port = invalid_port_number, got %s", config.Server.Port)
	}

	// Invalid integer should default to zero
	if config.Cache.DefaultExpireTimeSec != 0 {
		t.Errorf("Expected Cache.DefaultExpireTimeSec to be 0, got %d", config.Cache.DefaultExpireTimeSec)
	}
}

func TestAppConfigEmptyFile(t *testing.T) {
	// Test with empty env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	config, err := NewAppConfig(envFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config == nil {
		t.Error("Expected config but got nil")
		return
	}

	// All values should be zero values
	if config.App.Name != "" {
		t.Errorf("Expected App.Name to be empty, got %s", config.App.Name)
	}

	if config.Server.Port != "" {
		t.Errorf("Expected Server.Port to be empty, got %s", config.Server.Port)
	}

	if config.Cache.RedisDB != 0 {
		t.Errorf("Expected Cache.RedisDB to be 0, got %d", config.Cache.RedisDB)
	}
}
