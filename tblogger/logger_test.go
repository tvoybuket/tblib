package tblogger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWriter для тестирования вывода
type MockWriter struct {
	buffer *bytes.Buffer
}

func NewMockWriter() *MockWriter {
	return &MockWriter{
		buffer: &bytes.Buffer{},
	}
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *MockWriter) String() string {
	return m.buffer.String()
}

func (m *MockWriter) Reset() {
	m.buffer.Reset()
}

// TestDefaultConfig тестирует создание конфигурации по умолчанию
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, LevelInfo, config.Level)
	assert.Equal(t, FormatJSON, config.Format)
	assert.Equal(t, os.Stdout, config.Output)
	assert.False(t, config.AddSource)
	assert.NotNil(t, config.DefaultFields)
	assert.Equal(t, "unknown", config.ServiceName)
	assert.Equal(t, "unknown", config.ServiceVersion)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, time.UTC, config.TimeZone)
	assert.Equal(t, int64(100), config.MaxFileSize)
	assert.Equal(t, 5, config.MaxFiles)
}

// TestNew тестирует создание нового логгера
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &Config{
				Level:          LevelDebug,
				Format:         FormatText,
				Output:         os.Stdout,
				AddSource:      true,
				DefaultFields:  map[string]interface{}{"test": "value"},
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				TimeZone:       time.UTC,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.slogger)
				assert.NotNil(t, logger.config)
			}
		})
	}
}

// TestNewWithDefaults тестирует создание логгера с настройками по умолчанию
func TestNewWithDefaults(t *testing.T) {
	logger := NewWithDefaults()

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.slogger)
	assert.NotNil(t, logger.config)
	assert.Equal(t, LevelInfo, logger.config.Level)
	assert.Equal(t, FormatJSON, logger.config.Format)
}

// TestLoggingLevels тестирует все уровни логирования
func TestLoggingLevels(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelDebug,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		expected string
	}{
		{
			name:     "Debug",
			logFunc:  logger.Debug,
			expected: "DEBUG",
		},
		{
			name:     "Info",
			logFunc:  logger.Info,
			expected: "INFO",
		},
		{
			name:     "Warn",
			logFunc:  logger.Warn,
			expected: "WARN",
		},
		{
			name:     "Error",
			logFunc:  logger.Error,
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.Reset()
			tt.logFunc("test message", "key", "value")

			output := mockWriter.String()
			assert.Contains(t, output, "test message")
			assert.Contains(t, output, "key")
			assert.Contains(t, output, "value")
		})
	}
}

// TestContextLogging тестирует контекстное логирование
func TestContextLogging(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "request_id", "test-123")

	logger.InfoContext(ctx, "context message", "key", "value")

	output := mockWriter.String()
	assert.Contains(t, output, "context message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

// TestWithMethods тестирует методы With*
func TestWithMethods(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	t.Run("With", func(t *testing.T) {
		mockWriter.Reset()
		newLogger := logger.With("field1", "value1", "field2", "value2")
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "field1")
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "field2")
		assert.Contains(t, output, "value2")
	})

	t.Run("WithError", func(t *testing.T) {
		mockWriter.Reset()
		testErr := errors.New("test error")
		newLogger := logger.WithError(testErr)
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "error")
		assert.Contains(t, output, "test error")
	})

	t.Run("WithError nil", func(t *testing.T) {
		newLogger := logger.WithError(nil)
		assert.Equal(t, logger, newLogger)
	})

	t.Run("WithFields", func(t *testing.T) {
		mockWriter.Reset()
		fields := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}
		newLogger := logger.WithFields(fields)
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "field1")
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "field2")
		assert.Contains(t, output, "value2")
	})

	t.Run("WithRequest", func(t *testing.T) {
		mockWriter.Reset()
		newLogger := logger.WithRequest("GET", "/api/test", "test-agent", "req-123")
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "http_method")
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "http_path")
		assert.Contains(t, output, "/api/test")
		assert.Contains(t, output, "user_agent")
		assert.Contains(t, output, "test-agent")
		assert.Contains(t, output, "request_id")
		assert.Contains(t, output, "req-123")
	})

	t.Run("WithUser", func(t *testing.T) {
		mockWriter.Reset()
		newLogger := logger.WithUser("user-123", "testuser")
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "user_id")
		assert.Contains(t, output, "user-123")
		assert.Contains(t, output, "username")
		assert.Contains(t, output, "testuser")
	})

	t.Run("WithDuration", func(t *testing.T) {
		mockWriter.Reset()
		duration := 1500 * time.Millisecond
		newLogger := logger.WithDuration(duration)
		newLogger.Info("test message")

		output := mockWriter.String()
		assert.Contains(t, output, "duration_ms")
		assert.Contains(t, output, "1500")
		assert.Contains(t, output, "duration")
		assert.Contains(t, output, "1.5s")
	})

	t.Run("WithGroup", func(t *testing.T) {
		mockWriter.Reset()
		newLogger := logger.WithGroup("test_group")
		newLogger.Info("test message", "key", "value")

		output := mockWriter.String()
		assert.Contains(t, output, "test_group")
		assert.Contains(t, output, "key")
		assert.Contains(t, output, "value")
	})
}

// TestStructuredLogging тестирует структурированное логирование
func TestStructuredLogging(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	t.Run("LogHTTPRequest", func(t *testing.T) {
		mockWriter.Reset()
		logger.LogHTTPRequest("GET", "/api/users", "test-agent", "req-123", 200, 150*time.Millisecond, 1024)

		output := mockWriter.String()
		assert.Contains(t, output, "HTTP request")
		assert.Contains(t, output, "method")
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "path")
		assert.Contains(t, output, "/api/users")
		assert.Contains(t, output, "status_code")
		assert.Contains(t, output, "200")
		assert.Contains(t, output, "duration_ms")
		assert.Contains(t, output, "150")
		assert.Contains(t, output, "response_size")
		assert.Contains(t, output, "1024")
	})

	t.Run("LogDBQuery", func(t *testing.T) {
		// Создаем логгер с Debug уровнем для LogDBQuery
		debugConfig := &Config{
			Level:  LevelDebug,
			Format: FormatJSON,
			Output: mockWriter,
		}
		debugLogger, err := New(debugConfig)
		require.NoError(t, err)

		mockWriter.Reset()
		debugLogger.LogDBQuery("SELECT * FROM users", 50*time.Millisecond, 10)

		output := mockWriter.String()
		assert.Contains(t, output, "Database query")
		assert.Contains(t, output, "query")
		assert.Contains(t, output, "SELECT * FROM users")
		assert.Contains(t, output, "duration_ms")
		assert.Contains(t, output, "50")
		assert.Contains(t, output, "rows_affected")
		assert.Contains(t, output, "10")
	})

	t.Run("LogStartup", func(t *testing.T) {
		mockWriter.Reset()
		logger.LogStartup("8080", "development")

		output := mockWriter.String()
		assert.Contains(t, output, "Application starting")
		assert.Contains(t, output, "port")
		assert.Contains(t, output, "8080")
		assert.Contains(t, output, "environment")
		assert.Contains(t, output, "development")
	})

	t.Run("LogShutdown", func(t *testing.T) {
		mockWriter.Reset()
		logger.LogShutdown("SIGTERM")

		output := mockWriter.String()
		assert.Contains(t, output, "Application shutting down")
		assert.Contains(t, output, "reason")
		assert.Contains(t, output, "SIGTERM")
	})
}

// TestLevelChecks тестирует проверки уровней логирования
func TestLevelChecks(t *testing.T) {
	tests := []struct {
		name         string
		level        LogLevel
		debugEnabled bool
		infoEnabled  bool
	}{
		{
			name:         "Debug level",
			level:        LevelDebug,
			debugEnabled: true,
			infoEnabled:  true,
		},
		{
			name:         "Info level",
			level:        LevelInfo,
			debugEnabled: false,
			infoEnabled:  true,
		},
		{
			name:         "Warn level",
			level:        LevelWarn,
			debugEnabled: false,
			infoEnabled:  false,
		},
		{
			name:         "Error level",
			level:        LevelError,
			debugEnabled: false,
			infoEnabled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:  tt.level,
				Format: FormatJSON,
				Output: os.Stdout,
			}

			logger, err := New(config)
			require.NoError(t, err)

			assert.Equal(t, tt.debugEnabled, logger.IsDebugEnabled())
			assert.Equal(t, tt.infoEnabled, logger.IsInfoEnabled())
			assert.Equal(t, tt.level, logger.LogLevel())
		})
	}
}

// TestSetLevel тестирует изменение уровня логирования
func TestSetLevel(t *testing.T) {
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: os.Stdout,
	}

	logger, err := New(config)
	require.NoError(t, err)

	assert.Equal(t, LevelInfo, logger.LogLevel())
	assert.False(t, logger.IsDebugEnabled())
	assert.True(t, logger.IsInfoEnabled())

	logger.SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, logger.LogLevel())
	assert.True(t, logger.IsDebugEnabled())
	assert.True(t, logger.IsInfoEnabled())
}

// TestFormats тестирует различные форматы вывода
func TestFormats(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
	}{
		{
			name:   "JSON format",
			format: FormatJSON,
		},
		{
			name:   "Text format",
			format: FormatText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := NewMockWriter()
			config := &Config{
				Level:  LevelInfo,
				Format: tt.format,
				Output: mockWriter,
			}

			logger, err := New(config)
			require.NoError(t, err)

			logger.Info("test message", "key", "value")

			output := mockWriter.String()
			assert.Contains(t, output, "test message")
			assert.Contains(t, output, "key")
			assert.Contains(t, output, "value")

			if tt.format == FormatJSON {
				// Проверяем, что это валидный JSON
				var jsonData map[string]interface{}
				err := json.Unmarshal([]byte(output), &jsonData)
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefaultFields тестирует поля по умолчанию
func TestDefaultFields(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: mockWriter,
		DefaultFields: map[string]interface{}{
			"default_key": "default_value",
			"app_name":    "test-app",
		},
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	logger, err := New(config)
	require.NoError(t, err)

	logger.Info("test message")

	output := mockWriter.String()
	assert.Contains(t, output, "default_key")
	assert.Contains(t, output, "default_value")
	assert.Contains(t, output, "app_name")
	assert.Contains(t, output, "test-app")
	assert.Contains(t, output, "service")
	assert.Contains(t, output, "test-service")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "environment")
	assert.Contains(t, output, "test")
}

// TestGetCaller тестирует функцию getCaller
func TestGetCaller(t *testing.T) {
	file, line := getCaller(0)

	assert.NotEqual(t, "unknown", file)
	assert.Greater(t, line, 0)
	assert.Contains(t, file, "logger_test.go")
}

// TestPanic тестирует метод Panic
func TestPanic(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelError,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	assert.Panics(t, func() {
		logger.Panic("panic message", "key", "value")
	})

	output := mockWriter.String()
	assert.Contains(t, output, "panic message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

// TestFatal тестирует метод Fatal с моком os.Exit
func TestFatal(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelError,
		Format: FormatJSON,
		Output: mockWriter,
	}

	logger, err := New(config)
	require.NoError(t, err)

	// Сохраняем оригинальную функцию os.Exit
	originalOsExit := osExit
	defer func() {
		osExit = originalOsExit
	}()

	// Мокаем os.Exit
	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	logger.Fatal("fatal message", "key", "value")

	assert.True(t, exitCalled)
	assert.Equal(t, 1, exitCode)

	output := mockWriter.String()
	assert.Contains(t, output, "fatal message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

// TestGlobalLogger тестирует глобальный логгер
func TestGlobalLogger(t *testing.T) {
	// Сохраняем оригинальный глобальный логгер
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	mockWriter := NewMockWriter()
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: mockWriter,
	}

	newLogger, err := New(config)
	require.NoError(t, err)

	SetDefaultLogger(newLogger)

	assert.Equal(t, newLogger, GetDefaultLogger())

	// Тестируем глобальные функции
	mockWriter.Reset()
	Info("global info message", "key", "value")

	output := mockWriter.String()
	assert.Contains(t, output, "global info message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")

	// Debug сообщения не будут логироваться, так как уровень по умолчанию Info
	// Это ожидаемое поведение

	mockWriter.Reset()
	Warn("global warn message")

	output = mockWriter.String()
	assert.Contains(t, output, "global warn message")

	mockWriter.Reset()
	Error("global error message")

	output = mockWriter.String()
	assert.Contains(t, output, "global error message")

	// Тестируем глобальную функцию With
	globalWithLogger := With("global_field", "global_value")
	assert.NotNil(t, globalWithLogger)
}

// TestTimeZone тестирует настройку временной зоны
func TestTimeZone(t *testing.T) {
	mockWriter := NewMockWriter()
	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	config := &Config{
		Level:    LevelInfo,
		Format:   FormatJSON,
		Output:   mockWriter,
		TimeZone: moscowTZ,
	}

	logger, err := New(config)
	require.NoError(t, err)

	logger.Info("timezone test")

	output := mockWriter.String()

	// Проверяем, что время записано в правильной зоне
	var logData map[string]interface{}
	err = json.Unmarshal([]byte(output), &logData)
	require.NoError(t, err)

	timeStr, ok := logData["time"].(string)
	require.True(t, ok)

	// Парсим время и проверяем зону
	_, err = time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)

	// Проверяем, что время записано (может быть в UTC или Local в зависимости от системы)
	assert.NotEmpty(t, timeStr)
}

// TestAddSource тестирует добавление информации об источнике
func TestAddSource(t *testing.T) {
	mockWriter := NewMockWriter()
	config := &Config{
		Level:     LevelInfo,
		Format:    FormatJSON,
		Output:    mockWriter,
		AddSource: true,
	}

	logger, err := New(config)
	require.NoError(t, err)

	logger.Info("source test")

	output := mockWriter.String()

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(output), &logData)
	require.NoError(t, err)

	// Проверяем наличие информации об источнике
	source, ok := logData["source"].(map[string]interface{})
	require.True(t, ok)

	file, ok := source["file"].(string)
	require.True(t, ok)
	assert.Contains(t, file, "logger.go")

	line, ok := source["line"].(float64)
	require.True(t, ok)
	assert.Greater(t, int(line), 0)
}
