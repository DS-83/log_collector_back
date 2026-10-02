package server

import (
	"context"
	"log"
	"log_collect/config"
	"log_collect/internal"
	"log_collect/internal/application/usecase"
	"log_collect/internal/infrastructure/hasher"
	"log_collect/internal/infrastructure/repo/pgsql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpInternal "log_collect/internal/delivery/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	httpSrv       *http.Server
	db            *pgxpool.Pool
	apiKeyUseCase internal.ApiKeyUseCase
	userUseCase   internal.UserUseCase
	eventUseCase  internal.EventUseCase
}

func NewApp(cfg *config.Config) *App {
	baseCtx := context.Background()

	ctx, cancel := context.WithTimeout(baseCtx, 10*time.Second) // таймаут на подключение к БД, не связан с cfg.Context.Timeout
	defer cancel()

	db := NewPgPool(ctx, cfg)

	// Repo
	apiKeyRepo := pgsql.NewApiKeyRepo(db)
	userRepo := pgsql.NewUserRepo(db)
	sessionRepo := pgsql.NewSessionRepo(db)
	eventRepo := pgsql.NewEventRepo(db)

	// Hasher
	bcryptHasher := hasher.NewBcryptHasher(cfg.Hasher.Cost)

	return &App{
		httpSrv:       &http.Server{},
		db:            db,
		apiKeyUseCase: usecase.NewApiKeyUseCase(apiKeyRepo),
		userUseCase:   usecase.NewAuthUseCase(userRepo, sessionRepo, bcryptHasher, cfg.Session.TTL),
		eventUseCase:  usecase.NewEventUseCase(eventRepo),
	}
}

func (a *App) Run(cfg *config.Config) error {
	// Init gin handler
	r := gin.Default()

	r.Use(
		// Настройка CORS
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}),
		func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Context.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		},
	)

	// Set up http handlers
	// Middlwares
	authApiMiddleWare := httpInternal.NewAuthApiMiddleware(a.apiKeyUseCase)
	authUserMiddleware := httpInternal.NewAuthUserMiddleware(a.userUseCase, cfg.Cookie.Name)

	// User SignIn endpoints
	userauth := r.Group("/userauth")
	cookie := httpInternal.NewCookieConfig(
		cfg.Cookie.Name,
		cfg.Cookie.Path,
		cfg.Cookie.Domain,
		cfg.Cookie.Secure,
		cfg.Cookie.HTTPOnly,
	)
	httpInternal.RegisterRoutesUser(userauth, a.userUseCase, cfg.Session.TTL, cookie)

	// Event api endpoints
	eventsApi := r.Group("/api/v1/events", authApiMiddleWare)
	httpInternal.RegisterRouteEventApi(eventsApi, a.eventUseCase)

	// Event users endpoints
	user := r.Group("/events", authUserMiddleware)
	httpInternal.RegisterRouteEventUser(user, a.eventUseCase)

	// HTTP Server
	a.httpSrv = &http.Server{
		Addr:           ":" + cfg.App.Port,
		Handler:        r,
		ReadTimeout:    cfg.Context.Timeout,
		WriteTimeout:   cfg.Context.Timeout,
		MaxHeaderBytes: 1 << 20,
	}

	errCh := make(chan error, 1)
	go func() {
		// go run $(go env GOROOT)/src/crypto/tls/generate_cert.go --host=localhost
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// if err := a.httpSrv.ListenAndServeTLS("cert-xyz.pem", "key-xyz.pem"); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := a.httpSrv.Shutdown(ctx)
	a.db.Close()
	return err
}

// Init PG database
func NewPgPool(ctx context.Context, cfg *config.Config) *pgxpool.Pool {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("parse dsn: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Fatalf("create pool: %v", err)
	}

	// NewWithConfig соединение сразу не открывает, поэтому проверяем явно
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("ping: %v", err)
	}
	return pool
}
