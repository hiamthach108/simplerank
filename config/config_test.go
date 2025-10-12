package config

import (
	"os"
	"testing"
)

func TestNewAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() func() // setup function returns cleanup function
		wantErr bool
		verify  func(*testing.T, *AppConfig)
	}{
		{
			name: "load from environment variables",
			setup: func() func() {
				// Set environment variables
				os.Setenv("APP_NAME", "test-app")
				os.Setenv("APP_VERSION", "1.0.0")
				os.Setenv("HTTP_HOST", "localhost")
				os.Setenv("HTTP_PORT", "8080")
				os.Setenv("LOG_LEVEL", "debug")
				os.Setenv("CACHE_DEFAULT_EXPIRE_TIME_SEC", "3600")
				os.Setenv("CACHE_CLEANUP_INTERVAL_HOUR", "24")
				os.Setenv("REDIS_HOST", "localhost")
				os.Setenv("REDIS_PORT", "6379")
				os.Setenv("REDIS_PASSWORD", "secret123")
				os.Setenv("REDIS_DB", "0")

				return func() {
					// Cleanup
					os.Unsetenv("APP_NAME")
					os.Unsetenv("APP_VERSION")
					os.Unsetenv("HTTP_HOST")
					os.Unsetenv("HTTP_PORT")
					os.Unsetenv("LOG_LEVEL")
					os.Unsetenv("CACHE_DEFAULT_EXPIRE_TIME_SEC")
					os.Unsetenv("CACHE_CLEANUP_INTERVAL_HOUR")
					os.Unsetenv("REDIS_HOST")
					os.Unsetenv("REDIS_PORT")
					os.Unsetenv("REDIS_PASSWORD")
					os.Unsetenv("REDIS_DB")
				}
			},
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
				if config.Logger.Level != "debug" {
					t.Errorf("Expected Logger.Level = debug, got %s", config.Logger.Level)
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
			name: "no environment variables",
			setup: func() func() {
				return func() {} // No cleanup needed
			},
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			config, err := NewAppConfig()

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
	cleanup := func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("REDIS_HOST")
	}
	defer cleanup()

	// Only set some environment variables
	os.Setenv("APP_NAME", "partial-app")
	os.Setenv("HTTP_PORT", "9000")
	os.Setenv("REDIS_HOST", "redis-server")

	config, err := NewAppConfig()
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

func TestAppConfigInvalidEnvVars(t *testing.T) {
	// Test with invalid environment variable values
	cleanup := func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("CACHE_DEFAULT_EXPIRE_TIME_SEC")
	}
	defer cleanup()

	// Set invalid environment variable values
	os.Setenv("APP_NAME", "")                                // Empty value
	os.Setenv("HTTP_PORT", "invalid_port_number")           // Invalid port
	os.Setenv("CACHE_DEFAULT_EXPIRE_TIME_SEC", "not_a_number") // Invalid integer

	config, err := NewAppConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

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

func TestAppConfigWithDotEnvFile(t *testing.T) {
	// Test that .env file is loaded when present
	// Create a temporary .env file in current directory
	envContent := `APP_NAME=dotenv-app
HTTP_PORT=3000
LOG_LEVEL=warn`

	// Save current directory
	origDir, _ := os.Getwd()
	defer func() {
		os.Chdir(origDir)
		os.Remove(".env") // Clean up
	}()

	// Create temporary directory and change to it
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	config, err := NewAppConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config == nil {
		t.Error("Expected config but got nil")
		return
	}

	// Verify values from .env file are loaded
	if config.App.Name != "dotenv-app" {
		t.Errorf("Expected App.Name = dotenv-app, got %s", config.App.Name)
	}

	if config.Server.Port != "3000" {
		t.Errorf("Expected Server.Port = 3000, got %s", config.Server.Port)
	}

	if config.Logger.Level != "warn" {
		t.Errorf("Expected Logger.Level = warn, got %s", config.Logger.Level)
	}
}
