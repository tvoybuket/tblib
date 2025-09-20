package tblogger

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLogLevel тестирует уровни логирования
func TestLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{
			name:     "Debug level",
			level:    LevelDebug,
			expected: "DEBUG",
		},
		{
			name:     "Info level",
			level:    LevelInfo,
			expected: "INFO",
		},
		{
			name:     "Warn level",
			level:    LevelWarn,
			expected: "WARN",
		},
		{
			name:     "Error level",
			level:    LevelError,
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

// TestOutputFormat тестирует форматы вывода
func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   OutputFormat
		expected string
	}{
		{
			name:     "JSON format",
			format:   FormatJSON,
			expected: "json",
		},
		{
			name:     "Text format",
			format:   FormatText,
			expected: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.format))
		})
	}
}

// TestConfigValidation тестирует валидацию конфигурации
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Level:          LevelInfo,
				Format:         FormatJSON,
				Output:         os.Stdout,
				AddSource:      false,
				DefaultFields:  make(map[string]interface{}),
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				TimeZone:       time.UTC,
				MaxFileSize:    100,
				MaxFiles:       5,
			},
			expectError: false,
		},
		{
			name: "config with custom fields",
			config: &Config{
				Level:  LevelDebug,
				Format: FormatText,
				Output: os.Stderr,
				DefaultFields: map[string]interface{}{
					"custom_field": "custom_value",
					"app_id":       "test-app-123",
				},
				ServiceName:    "custom-service",
				ServiceVersion: "2.0.0",
				Environment:    "production",
				TimeZone:       time.UTC,
				MaxFileSize:    200,
				MaxFiles:       10,
			},
			expectError: false,
		},
		{
			name: "config with file path",
			config: &Config{
				Level:    LevelInfo,
				Format:   FormatJSON,
				FilePath: "/tmp/test.log",
				TimeZone: time.UTC,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.Equal(t, tt.config.Level, logger.config.Level)
				assert.Equal(t, tt.config.Format, logger.config.Format)
				assert.Equal(t, tt.config.ServiceName, logger.config.ServiceName)
				assert.Equal(t, tt.config.ServiceVersion, logger.config.ServiceVersion)
				assert.Equal(t, tt.config.Environment, logger.config.Environment)
			}
		})
	}
}

// TestConfigDefaults тестирует значения по умолчанию
func TestConfigDefaults(t *testing.T) {
	config := DefaultConfig()

	// Проверяем все поля конфигурации по умолчанию
	assert.Equal(t, LevelInfo, config.Level)
	assert.Equal(t, FormatJSON, config.Format)
	assert.Equal(t, os.Stdout, config.Output)
	assert.False(t, config.AddSource)
	assert.NotNil(t, config.DefaultFields)
	assert.Empty(t, config.DefaultFields)
	assert.Equal(t, "unknown", config.ServiceName)
	assert.Equal(t, "unknown", config.ServiceVersion)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, time.UTC, config.TimeZone)
	assert.Equal(t, int64(100), config.MaxFileSize)
	assert.Equal(t, 5, config.MaxFiles)
	assert.Empty(t, config.FilePath)
}

// TestConfigCopy тестирует копирование конфигурации
func TestConfigCopy(t *testing.T) {
	original := &Config{
		Level:  LevelDebug,
		Format: FormatText,
		Output: os.Stderr,
		DefaultFields: map[string]interface{}{
			"test": "value",
		},
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		TimeZone:       time.UTC,
		MaxFileSize:    150,
		MaxFiles:       7,
	}

	// Создаем копию конфигурации
	copy := *original

	// Создаем новую map для DefaultFields
	copy.DefaultFields = make(map[string]interface{})
	for k, v := range original.DefaultFields {
		copy.DefaultFields[k] = v
	}

	// Изменяем копию
	copy.Level = LevelError
	copy.ServiceName = "modified-service"
	copy.DefaultFields["new_field"] = "new_value"

	// Проверяем, что оригинал не изменился
	assert.Equal(t, LevelDebug, original.Level)
	assert.Equal(t, "test-service", original.ServiceName)
	assert.Len(t, original.DefaultFields, 1)
	assert.Equal(t, "value", original.DefaultFields["test"])

	// Проверяем, что копия изменилась
	assert.Equal(t, LevelError, copy.Level)
	assert.Equal(t, "modified-service", copy.ServiceName)
	assert.Len(t, copy.DefaultFields, 2)
	assert.Equal(t, "value", copy.DefaultFields["test"])
	assert.Equal(t, "new_value", copy.DefaultFields["new_field"])
}

// TestConfigWithNilFields тестирует конфигурацию с nil полями
func TestConfigWithNilFields(t *testing.T) {
	config := &Config{
		Level:         LevelInfo,
		Format:        FormatJSON,
		Output:        os.Stdout,
		DefaultFields: nil, // nil поля
		TimeZone:      nil, // nil временная зона
	}

	logger, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Проверяем, что nil поля обрабатываются корректно
	assert.Nil(t, logger.config.DefaultFields)
	assert.Nil(t, logger.config.TimeZone)
}

// TestConfigTimeZone тестирует различные временные зоны
func TestConfigTimeZone(t *testing.T) {
	timeZones := []*time.Location{
		time.UTC,
		nil, // nil зона
	}

	for _, tz := range timeZones {
		t.Run("timezone_"+getTimeZoneName(tz), func(t *testing.T) {
			config := &Config{
				Level:    LevelInfo,
				Format:   FormatJSON,
				Output:   os.Stdout,
				TimeZone: tz,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tz, logger.config.TimeZone)
		})
	}
}

// TestConfigMaxFileSize тестирует различные размеры файлов
func TestConfigMaxFileSize(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
	}{
		{
			name:     "small file size",
			fileSize: 1,
		},
		{
			name:     "medium file size",
			fileSize: 100,
		},
		{
			name:     "large file size",
			fileSize: 1000,
		},
		{
			name:     "zero file size",
			fileSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:       LevelInfo,
				Format:      FormatJSON,
				Output:      os.Stdout,
				MaxFileSize: tt.fileSize,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.fileSize, logger.config.MaxFileSize)
		})
	}
}

// TestConfigMaxFiles тестирует различное количество файлов
func TestConfigMaxFiles(t *testing.T) {
	tests := []struct {
		name     string
		maxFiles int
	}{
		{
			name:     "single file",
			maxFiles: 1,
		},
		{
			name:     "multiple files",
			maxFiles: 5,
		},
		{
			name:     "many files",
			maxFiles: 100,
		},
		{
			name:     "zero files",
			maxFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:    LevelInfo,
				Format:   FormatJSON,
				Output:   os.Stdout,
				MaxFiles: tt.maxFiles,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.maxFiles, logger.config.MaxFiles)
		})
	}
}

// TestConfigServiceInfo тестирует информацию о сервисе
func TestConfigServiceInfo(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		serviceVersion string
		environment    string
	}{
		{
			name:           "development environment",
			serviceName:    "my-service",
			serviceVersion: "1.0.0",
			environment:    "development",
		},
		{
			name:           "production environment",
			serviceName:    "production-service",
			serviceVersion: "2.1.0",
			environment:    "production",
		},
		{
			name:           "staging environment",
			serviceName:    "staging-service",
			serviceVersion: "2.0.0-beta",
			environment:    "staging",
		},
		{
			name:           "empty values",
			serviceName:    "",
			serviceVersion: "",
			environment:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:          LevelInfo,
				Format:         FormatJSON,
				Output:         os.Stdout,
				ServiceName:    tt.serviceName,
				ServiceVersion: tt.serviceVersion,
				Environment:    tt.environment,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.serviceName, logger.config.ServiceName)
			assert.Equal(t, tt.serviceVersion, logger.config.ServiceVersion)
			assert.Equal(t, tt.environment, logger.config.Environment)
		})
	}
}

// TestConfigDefaultFields тестирует поля по умолчанию
func TestConfigDefaultFields(t *testing.T) {
	tests := []struct {
		name          string
		defaultFields map[string]interface{}
	}{
		{
			name:          "empty fields",
			defaultFields: map[string]interface{}{},
		},
		{
			name: "single field",
			defaultFields: map[string]interface{}{
				"app_id": "test-app-123",
			},
		},
		{
			name: "multiple fields",
			defaultFields: map[string]interface{}{
				"app_id":  "test-app-123",
				"region":  "us-east-1",
				"cluster": "production",
				"node_id": "node-001",
			},
		},
		{
			name: "fields with different types",
			defaultFields: map[string]interface{}{
				"string_field": "string_value",
				"int_field":    42,
				"bool_field":   true,
				"float_field":  3.14,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:         LevelInfo,
				Format:        FormatJSON,
				Output:        os.Stdout,
				DefaultFields: tt.defaultFields,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.defaultFields, logger.config.DefaultFields)
		})
	}
}

// TestConfigAddSource тестирует настройку AddSource
func TestConfigAddSource(t *testing.T) {
	tests := []struct {
		name      string
		addSource bool
	}{
		{
			name:      "add source enabled",
			addSource: true,
		},
		{
			name:      "add source disabled",
			addSource: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:     LevelInfo,
				Format:    FormatJSON,
				Output:    os.Stdout,
				AddSource: tt.addSource,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.addSource, logger.config.AddSource)
		})
	}
}

// TestConfigOutput тестирует различные типы вывода
func TestConfigOutput(t *testing.T) {
	tests := []struct {
		name   string
		output io.Writer
	}{
		{
			name:   "stdout",
			output: os.Stdout,
		},
		{
			name:   "stderr",
			output: os.Stderr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:  LevelInfo,
				Format: FormatJSON,
				Output: tt.output,
			}

			logger, err := New(config)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.output, logger.config.Output)
		})
	}
}

// Вспомогательная функция для получения имени временной зоны
func getTimeZoneName(tz *time.Location) string {
	if tz == nil {
		return "nil"
	}
	return tz.String()
}
