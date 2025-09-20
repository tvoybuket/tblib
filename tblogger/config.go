package tblogger

import (
	"io"
	"log/slog"
	"time"
)

// LogLevel представляет уровень логирования
type LogLevel slog.Level

// Предопределенные уровни логирования
const (
	LevelDebug LogLevel = LogLevel(slog.LevelDebug)
	LevelInfo  LogLevel = LogLevel(slog.LevelInfo)
	LevelWarn  LogLevel = LogLevel(slog.LevelWarn)
	LevelError LogLevel = LogLevel(slog.LevelError)
)

// String возвращает строковое представление уровня логирования
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// OutputFormat определяет формат вывода логов
type OutputFormat string

const (
	FormatJSON OutputFormat = "json"
	FormatText OutputFormat = "text"
)

// Config содержит конфигурацию логгера
type Config struct {
	// Уровень логирования
	Level LogLevel

	// Формат вывода (json/text)
	Format OutputFormat

	// Вывод логов (файл или stdout/stderr)
	Output io.Writer

	// Путь к файлу логов (если нужно писать в файл)
	FilePath string

	// Максимальный размер файла в MB (для ротации)
	MaxFileSize int64

	// Количество файлов для ротации
	MaxFiles int

	// Включать ли информацию о коде (файл, строка)
	AddSource bool

	// Кастомные поля, добавляемые ко всем логам
	DefaultFields map[string]interface{}

	// Имя сервиса
	ServiceName string

	// Версия сервиса
	ServiceVersion string

	// Окружение (dev, staging, prod)
	Environment string

	// Временная зона
	TimeZone *time.Location
}
