package config

type Config struct {
	Logger Log
	Auth   Auth
	DB     DB
	HTTP   HTTP
}

type Log struct {
	Application string `env:"LOG_APPLICATION"`
	File        string `env:"LOG_FILE"`
	Level       string `env:"LOG_LEVEL"`
	AddSource   bool   `env:"LOG_ADD_SOURCE"`
}
type DB struct {
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Driver   string `env:"DB_DRIVER"`
	Name     string `env:"DB_NAME"`
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
}

type Auth struct {
	OIDCIssuer          string
	OIDCClientID        string
	OIDCClientSecret    string
	OIDCRedirectURL     string
	OIDCSuccessRedirect string
}

type HTTP struct {
	Host       string `env:"HOST"`
	Port       string `env:"PORT"`
	ExposePort string `env:"EXPOSE_PORT"`
}
