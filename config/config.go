package config

import (
	e "log_collect/err"

	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/spf13/viper"
)

// config keys
const (
	// App port
	appPort = "app.port"

	// Context
	ctxTimeout = "context.timeout"

	// Session
	sessionTTL = "session.ttl"

	// Cookie
	cookieName   = "cookie.name"
	cookiePath   = "cookie.path"
	cookieDomain = "cookie.domain"
	cookieSec    = "cookie.secure"
	cookieHttp   = "cookie.http_only"

	// Hasher
	hasherCost = "hasher.cost"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	Context ContextConfig
	Session SessionConfig
	Cookie  CookieConfig
	Hasher  HasherConfig
}

type HasherConfig struct {
	Cost int
}

type AppConfig struct {
	Port string
}

type DBConfig struct {
	DSN string
}

type ContextConfig struct {
	Timeout time.Duration
}

type SessionConfig struct {
	TTL time.Duration
}

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	HTTPOnly bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	env := os.Getenv("APP_ENV")

	configFile := "./config/config.dev.yml"
	if env == "production" {
		configFile = "./config/config.prod.yml"
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// DB
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, e.ErrDbDsnNotSet
	}

	// Session TTL
	sessionTTL := viper.GetDuration(sessionTTL)
	if sessionTTL <= 0 {
		sessionTTL = 10 * time.Minute // default
	}
	// Contex timeout
	ctxTimeout := viper.GetDuration(ctxTimeout)
	if ctxTimeout <= 0 {
		ctxTimeout = 10 * time.Second // default
	}

	cfg := &Config{
		App: AppConfig{
			Port: viper.GetString(appPort),
		},
		DB: DBConfig{
			DSN: dsn,
		},
		Context: ContextConfig{
			Timeout: ctxTimeout,
		},
		Session: SessionConfig{
			TTL: sessionTTL,
		},
		Cookie: CookieConfig{
			Name:     viper.GetString(cookieName),
			Path:     viper.GetString(cookiePath),
			Domain:   viper.GetString(cookieDomain),
			Secure:   viper.GetBool(cookieSec),
			HTTPOnly: viper.GetBool(cookieHttp),
		},
		Hasher: HasherConfig{
			Cost: viper.GetInt(hasherCost),
		},
	}
	return cfg, nil
}
