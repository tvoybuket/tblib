package tblogger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger представляет настроенный логгер
type Logger struct {
	slogger *slog.Logger
	config  *Config
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Level:          LevelInfo,
		Format:         FormatJSON,
		Output:         os.Stdout,
		AddSource:      false,
		DefaultFields:  make(map[string]interface{}),
		ServiceName:    "unknown",
		ServiceVersion: "unknown",
		Environment:    "development",
		TimeZone:       time.UTC,
		MaxFileSize:    100, // 100MB
		MaxFiles:       5,
	}
}

// New создает новый логгер с указанной конфигурацией
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Настройка вывода
	var output io.Writer = config.Output
	if config.FilePath != "" {
		file, err := setupFileOutput(config.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to setup file output: %w", err)
		}
		output = file
	}

	// Создание обработчика в зависимости от формата
	var handler slog.Handler
	handlerOptions := &slog.HandlerOptions{
		Level:     slog.Level(config.Level),
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Кастомизация атрибутов времени
			if a.Key == slog.TimeKey {
				if config.TimeZone != nil {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.TimeValue(a.Value.Time().In(config.TimeZone)),
					}
				}
			}
			return a
		},
	}

	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, handlerOptions)
	case FormatText:
		handler = slog.NewTextHandler(output, handlerOptions)
	default:
		handler = slog.NewJSONHandler(output, handlerOptions)
	}

	// Создание slog logger
	slogger := slog.New(handler)

	// Добавление контекстных полей по умолчанию
	contextFields := []interface{}{
		"service", config.ServiceName,
		"version", config.ServiceVersion,
		"environment", config.Environment,
	}

	// Добавление кастомных полей по умолчанию
	for key, value := range config.DefaultFields {
		contextFields = append(contextFields, key, value)
	}

	if len(contextFields) > 0 {
		slogger = slogger.With(contextFields...)
	}

	return &Logger{
		slogger: slogger,
		config:  config,
	}, nil
}

// NewWithDefaults создает логгер с настройками по умолчанию
func NewWithDefaults() *Logger {
	logger, _ := New(DefaultConfig())
	return logger
}

// Debug логирует сообщение на уровне DEBUG
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.slogger.Debug(msg, args...)
}

// DebugContext логирует сообщение на уровне DEBUG с контекстом
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.slogger.DebugContext(ctx, msg, args...)
}

// Info логирует сообщение на уровне INFO
func (l *Logger) Info(msg string, args ...interface{}) {
	l.slogger.Info(msg, args...)
}

// InfoContext логирует сообщение на уровне INFO с контекстом
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.slogger.InfoContext(ctx, msg, args...)
}

// Warn логирует сообщение на уровне WARN
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.slogger.Warn(msg, args...)
}

// WarnContext логирует сообщение на уровне WARN с контекстом
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.slogger.WarnContext(ctx, msg, args...)
}

// Error логирует сообщение на уровне ERROR
func (l *Logger) Error(msg string, args ...interface{}) {
	l.slogger.Error(msg, args...)
}

// ErrorContext логирует сообщение на уровне ERROR с контекстом
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.slogger.ErrorContext(ctx, msg, args...)
}

// With возвращает новый логгер с дополнительными полями
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		slogger: l.slogger.With(args...),
		config:  l.config,
	}
}

// WithGroup возвращает новый логгер с группировкой полей
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		slogger: l.slogger.WithGroup(name),
		config:  l.config,
	}
}

// WithError добавляет информацию об ошибке в лог
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.With("error", err.Error())
}

// WithFields добавляет несколько полей одновременно
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}
	return l.With(args...)
}

// WithRequest добавляет информацию о HTTP запросе
func (l *Logger) WithRequest(method, path, userAgent, requestID string) *Logger {
	return l.With(
		"http_method", method,
		"http_path", path,
		"user_agent", userAgent,
		"request_id", requestID,
	)
}

// WithUser добавляет информацию о пользователе
func (l *Logger) WithUser(userID, username string) *Logger {
	return l.With(
		"user_id", userID,
		"username", username,
	)
}

// WithDuration добавляет информацию о длительности операции
func (l *Logger) WithDuration(duration time.Duration) *Logger {
	return l.With(
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
}

// LogLevel возвращает текущий уровень логирования
func (l *Logger) LogLevel() LogLevel {
	return l.config.Level
}

// SetLevel изменяет уровень логирования
func (l *Logger) SetLevel(level LogLevel) {
	l.config.Level = level
}

// IsDebugEnabled проверяет, включен ли уровень DEBUG
func (l *Logger) IsDebugEnabled() bool {
	return l.config.Level <= LevelDebug
}

// IsInfoEnabled проверяет, включен ли уровень INFO
func (l *Logger) IsInfoEnabled() bool {
	return l.config.Level <= LevelInfo
}

// Метод для получения информации о вызывающем коде
func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", 0
	}
	return filepath.Base(file), line
}

// setupFileOutput настраивает вывод в файл
func setupFileOutput(filePath string) (io.Writer, error) {
	// Создание директории если не существует
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Открытие файла для записи
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

// Fatal логирует сообщение на уровне ERROR и завершает программу
var osExit = os.Exit

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.slogger.Error(msg, args...)
	osExit(1)
}

// Panic логирует сообщение на уровне ERROR и вызывает panic
func (l *Logger) Panic(msg string, args ...interface{}) {
	l.slogger.Error(msg, args...)
	panic(msg)
}

// Structured logging helpers

// LogHTTPRequest логирует HTTP запрос
func (l *Logger) LogHTTPRequest(method, path, userAgent, requestID string, statusCode int, duration time.Duration, size int64) {
	l.Info("HTTP request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"response_size", size,
		"user_agent", userAgent,
		"request_id", requestID,
	)
}

// LogDBQuery логирует запрос к базе данных
func (l *Logger) LogDBQuery(query string, duration time.Duration, rowsAffected int64) {
	l.Debug("Database query",
		"query", query,
		"duration_ms", duration.Milliseconds(),
		"rows_affected", rowsAffected,
	)
}

// LogStartup логирует запуск приложения
func (l *Logger) LogStartup(port string, env string) {
	l.Info("Application starting",
		"port", port,
		"environment", env,
		"service", l.config.ServiceName,
		"version", l.config.ServiceVersion,
	)
}

// LogShutdown логирует остановку приложения
func (l *Logger) LogShutdown(reason string) {
	l.Info("Application shutting down",
		"reason", reason,
		"service", l.config.ServiceName,
	)
}

// Global logger instance for convenience
var defaultLogger *Logger

func init() {
	defaultLogger = NewWithDefaults()
}

// Global functions that use the default logger

// SetDefaultLogger устанавливает глобальный логгер по умолчанию
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

// GetDefaultLogger возвращает глобальный логгер по умолчанию
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// Debug использует глобальный логгер
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info использует глобальный логгер
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn использует глобальный логгер
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error использует глобальный логгер
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// Fatal использует глобальный логгер
func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatal(msg, args...)
}

// With возвращает новый логгер с дополнительными полями (глобальный)
func With(args ...interface{}) *Logger {
	return defaultLogger.With(args...)
}
