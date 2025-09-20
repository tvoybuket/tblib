package tbconfig

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// ConfigField определяет конфигурацию для каждого поля
type ConfigField struct {
	EnvVar      string        // Имя переменной окружения
	Default     string        // Значение по умолчанию в виде строки
	Required    bool          // Является ли поле обязательным
	Transform   TransformType // Возможные значения: TransformURLEscape, TransformHostsNoPorts
	Separator   string        // Для слайсов, какой разделитель использовать
	Description string        // Описание
}

type Env string

const (
	EnvProd    = Env("production")
	EnvStaging = Env("staging")
	EnvLocal   = Env("local")
)

const ServiceEnvVarName = "NODE_ENV"

// Константы для Transform
const (
	TransformURLEscape    TransformType = "url_escape"     // Преобразует строку в URL-кодированную форму
	TransformHostsNoPorts TransformType = "hosts_no_ports" // Убирает порты из строки хостов
)

// Константы для парсинга тегов config
const (
	configTagEnv       = "env:"
	configTagDefault   = "default:"
	configTagSep       = "sep:"
	configTagTransform = "transform:"
	configTagDesc      = "desc:"
	configTagRequired  = "required"
)

// Тип для Transform как enum
// TransformType определяет допустимые значения для поля Transform
// Используйте только значения, определённые ниже

type TransformType string

// LoadConfig загружает конфигурацию в любую структуру с тегами config
func LoadConfig(cfg interface{}) error {
	env := getEnv(ServiceEnvVarName, string(EnvLocal))
	if env == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}

	if err := loadConfigIntoStruct(cfg); err != nil {
		return err
	}

	// Установить поле Env, если оно существует
	v := reflect.ValueOf(cfg).Elem()
	if envField := v.FieldByName("Env"); envField.IsValid() && envField.CanSet() {
		if envField.Kind() == reflect.String {
			envField.SetString(env)
		} else if envField.Type().Name() == "Env" {
			envField.Set(reflect.ValueOf(Env(env)))
		}
	}

	return nil
}

// loadConfigIntoStruct использует рефлексию для загрузки конфигурации из тегов структуры
func loadConfigIntoStruct(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("config")
		if tag == "" {
			continue
		}

		config, err := parseConfigTag(tag)
		if err != nil {
			return fmt.Errorf("error parsing config tag for field %s: %v", fieldType.Name, err)
		}

		if err := setFieldValue(field, config); err != nil {
			return fmt.Errorf("error setting field %s: %v", fieldType.Name, err)
		}
	}

	return nil
}

// parseConfigTag разбирает тег структуры config
func parseConfigTag(tag string) (*ConfigField, error) {
	config := &ConfigField{}
	parts := strings.SplitSeq(tag, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, configTagEnv):
			config.EnvVar = strings.TrimPrefix(part, configTagEnv)
		case strings.HasPrefix(part, configTagDefault):
			config.Default = strings.TrimPrefix(part, configTagDefault)
		case strings.HasPrefix(part, configTagSep):
			sep := strings.TrimPrefix(part, configTagSep)
			if strings.HasPrefix(sep, "'") && strings.HasSuffix(sep, "'") {
				sep = strings.Trim(sep, "'")
			}
			config.Separator = sep
		case strings.HasPrefix(part, configTagTransform):
			config.Transform = TransformType(strings.TrimPrefix(part, configTagTransform))
		case strings.HasPrefix(part, configTagDesc):
			config.Description = strings.TrimPrefix(part, configTagDesc)
		case part == configTagRequired:
			config.Required = true
		}
	}

	return config, nil
}

// setFieldValue устанавливает значение поля на основе его типа и конфигурации
func setFieldValue(field reflect.Value, config *ConfigField) error {
	rawValue := getEnv(config.EnvVar, config.Default)

	if config.Required && rawValue == "" {
		return fmt.Errorf("обязательная переменная окружения %s не установлена", config.EnvVar)
	}

	transformedValue := applyTransform(rawValue, config.Transform)

	switch field.Kind() {
	case reflect.String:
		field.SetString(transformedValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if transformedValue == "" {
			return nil
		}
		intVal, err := strconv.ParseInt(transformedValue, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %s as int: %v", transformedValue, err)
		}
		field.SetInt(intVal)
	case reflect.Bool:
		if transformedValue == "" {
			return nil
		}
		boolVal, err := strconv.ParseBool(transformedValue)
		if err != nil {
			return fmt.Errorf("cannot parse %s as bool: %v", transformedValue, err)
		}
		field.SetBool(boolVal)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			separator := config.Separator
			if separator == "" {
				separator = ","
			}

			var slice []string
			if transformedValue != "" {
				slice = strings.Split(transformedValue, separator)
				if config.Transform == TransformHostsNoPorts {
					for i, host := range slice {
						slice[i] = strings.Split(host, ":")[0]
					}
				}
			} else if config.Default != "" {
				slice = strings.Split(config.Default, separator)
			}

			field.Set(reflect.ValueOf(slice))
		} else {
			return fmt.Errorf("unsupported slice type: %v", field.Type())
		}
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}

// applyTransform применяет указанное преобразование к значению
func applyTransform(value string, transform TransformType) string {
	switch transform {
	case TransformURLEscape:
		return url.QueryEscape(value)
	case TransformHostsNoPorts:
		return value
	default:
		return value
	}
}

// Вспомогательные функции
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func GetEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func GetEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func GetEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	return strings.Split(valStr, sep)
}
