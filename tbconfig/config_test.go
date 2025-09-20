package tbconfig

import (
	"os"
	"reflect"
	"testing"
)

type Settings struct {
	ScyllaHosts    []string `config:"env:SCYLLA_CONTACT_POINTS,sep:',',transform:hosts_no_ports"`
	ScyllaUsername string   `config:"env:SCYLLA_USERNAME"`
	ScyllaPassword string   `config:"env:SCYLLA_PASSWORD"`
	ScyllaDC       string   `config:"env:SCYLLA_DC"`

	RabbitHost          string `config:"env:RABBIT_HOST"`
	RabbitPassword      string `config:"env:RABBIT_PASSWORD,transform:url_escape"`
	RabbitPort          string `config:"env:RABBIT_PORT"`
	RabbitUser          string `config:"env:RABBIT_USER"`
	RabbitConnectionUrl string

	DbHost          string `config:"env:DB_HOST"`
	DbName          string `config:"env:DB_NAME"`
	DbPassword      string `config:"env:DB_SESSION_PASSWORD,transform:url_escape"`
	DbPort          string `config:"env:DB_PORT"`
	DbUsername      string `config:"env:DB_SESSION_USERNAME"`
	DbConnectionUrl string

	ApiSmartChatUrl string `config:"env:API_SMART_CHAT_URL"`
	Port            int
	SmartChatEmail  string
	WpUrl           string `config:"env:WP_URL"`
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("SCYLLA_CONTACT_POINTS", "host1:9042,host2:9042,host3:9042")
	os.Setenv("SCYLLA_USERNAME", "user1")
	os.Setenv("SCYLLA_PASSWORD", "pass1")
	os.Setenv("SCYLLA_DC", "dc1")
	os.Setenv("RABBIT_HOST", "rabbit.local")
	os.Setenv("RABBIT_PASSWORD", "p@ss w!th sp&cial")
	os.Setenv("RABBIT_PORT", "5672")
	os.Setenv("RABBIT_USER", "rabbituser")
	os.Setenv("DB_HOST", "db.local")
	os.Setenv("DB_NAME", "dbname")
	os.Setenv("DB_SESSION_PASSWORD", "db@pass w!th sp&cial")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_SESSION_USERNAME", "dbuser")
	os.Setenv("API_SMART_CHAT_URL", "https://api.smartchat.local")
	os.Setenv("WP_URL", "https://wp.local")
	os.Setenv("NODE_ENV", "test")

	settings := &Settings{}
	err := LoadConfig(settings)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expectedHosts := []string{"host1", "host2", "host3"}
	if !reflect.DeepEqual(settings.ScyllaHosts, expectedHosts) {
		t.Errorf("ScyllaHosts mismatch: got %v, want %v", settings.ScyllaHosts, expectedHosts)
	}

	if settings.ScyllaUsername != "user1" {
		t.Errorf("ScyllaUsername mismatch: got %s, want %s", settings.ScyllaUsername, "user1")
	}
	if settings.ScyllaPassword != "pass1" {
		t.Errorf("ScyllaPassword mismatch: got %s, want %s", settings.ScyllaPassword, "pass1")
	}
	if settings.ScyllaDC != "dc1" {
		t.Errorf("ScyllaDC mismatch: got %s, want %s", settings.ScyllaDC, "dc1")
	}
	if settings.RabbitHost != "rabbit.local" {
		t.Errorf("RabbitHost mismatch: got %s, want %s", settings.RabbitHost, "rabbit.local")
	}
	if settings.RabbitPassword != "p%40ss+w%21th+sp%26cial" {
		t.Errorf("RabbitPassword mismatch: got %s, want %s", settings.RabbitPassword, "p%40ss+w%21th+sp%26cial")
	}
	if settings.RabbitPort != "5672" {
		t.Errorf("RabbitPort mismatch: got %s, want %s", settings.RabbitPort, "5672")
	}
	if settings.RabbitUser != "rabbituser" {
		t.Errorf("RabbitUser mismatch: got %s, want %s", settings.RabbitUser, "rabbituser")
	}
	if settings.DbHost != "db.local" {
		t.Errorf("DbHost mismatch: got %s, want %s", settings.DbHost, "db.local")
	}
	if settings.DbName != "dbname" {
		t.Errorf("DbName mismatch: got %s, want %s", settings.DbName, "dbname")
	}
	if settings.DbPassword != "db%40pass+w%21th+sp%26cial" {
		t.Errorf("DbPassword mismatch: got %s, want %s", settings.DbPassword, "db%40pass+w%21th+sp%26cial")
	}
	if settings.DbPort != "5432" {
		t.Errorf("DbPort mismatch: got %s, want %s", settings.DbPort, "5432")
	}
	if settings.DbUsername != "dbuser" {
		t.Errorf("DbUsername mismatch: got %s, want %s", settings.DbUsername, "dbuser")
	}
	if settings.ApiSmartChatUrl != "https://api.smartchat.local" {
		t.Errorf("ApiSmartChatUrl mismatch: got %s, want %s", settings.ApiSmartChatUrl, "https://api.smartchat.local")
	}
	if settings.WpUrl != "https://wp.local" {
		t.Errorf("WpUrl mismatch: got %s, want %s", settings.WpUrl, "https://wp.local")
	}
}
