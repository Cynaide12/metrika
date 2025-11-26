package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
    Env                       string     `yaml:"env" env-default:"local" env-required:"true" env:"ENV"`
    LogFilePath               string     `yaml:"log_file_path" env-required:"true" env-default:".logs/apm_server.log" env:"LOG_FILE_PATH"`
    JWTSecret                 string     `yaml:"jwt_secret" env-required:"true" env:"JWT_SECRET"`
    MaxRequestSize            int        `yaml:"max_request_size" env-default:"64" env:"MAX_REQUEST_SIZE"`
    FilesStoragePath          string     `yaml:"files_storage_path" env-default:"./var/uploads" env:"FILES_STORAGE_PATH"`
    AllowedFilesExtensionsRaw string     `yaml:"allowed_file_extensions" env:"ALLOWED_FILES_EXTENSIONS_RAW"`
    AllowedFilesExtension     []string   //парсится в MustLoad
    MaxFileSize               int64      `yaml:"max_file_size" env-default:"20" env:"MAX_FILE_SIZE"`
    HTTPServer                HTTPServer `yaml:"http_server"`
    DBServer                  DBServer   `yaml:"db_server"`
    // SMTPServer                SMTPServer `yaml:"smtp_server"`
    Frontend                  Frontend   `yaml:"frontend"`
}

type HTTPServer struct {
    Address     string        `yaml:"address" env-default:"localhost:8080" env:"HTTP_SERVER_ADDRESS"`
    Timeout     time.Duration `yaml:"timeout" env-default:"4s" env:"HTTP_SERVER_TIMEOUT"`
    IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s" env:"HTTP_SERVER_IDLE_TIMEOUT"`
}

type DBServer struct {
    Host     string `yaml:"host" env-required:"true" env-default:"localhost" env:"DB_HOST"`
    Port     int    `yaml:"port" env-required:"true" env-default:"5432" env:"DB_PORT"`
    Username string `yaml:"username" env-required:"true" env-default:"postgres" env:"DB_USERNAME"`
    Password string `yaml:"password" env-required:"true" env-default:"root" env:"DB_PASSWORD"`
    DBName   string `yaml:"db_name" env-required:"true" env:"DB_NAME"`
}

// type SMTPServer struct {
//     Host     string `yaml:"host" env-required:"true" env:"SMTP_HOST"`
//     Port     int    `yaml:"port" env-required:"true" env:"SMTP_PORT"`
//     Username string `yaml:"username" env-required:"true" env:"SMTP_USERNAME"`
//     Password string `yaml:"password" env-required:"true" env:"SMTP_PASSWORD"`
//     Sender   string `yaml:"sender" env-required:"true" env:"SMTP_SENDER"`
// }

type Frontend struct {
    AppUrl string `yaml:"app_url" env-required:"true" env:"FRONTEND_APP_URL"`
}


func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	//перекрываем значение из yaml переменными из env(чтобы можно было, например, в докере указать env)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}

	//парсим разрешенные типы файлов
	if cfg.AllowedFilesExtensionsRaw != "" {
		allowedExtensions := strings.Split(cfg.AllowedFilesExtensionsRaw, ";")
		if len(allowedExtensions) > 0 {
			cfg.AllowedFilesExtension = allowedExtensions
		}
	}

	return &cfg
}
