package main

import (
	"billify-api/internal/auth"
	"billify-api/internal/db"
	"billify-api/internal/env"
	"billify-api/internal/pdf"
	"billify-api/internal/store"
	"expvar"
	"log"
	"runtime"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const version = "1.0.0"

func main() {

    cfg := config{
        addr:        env.GetString("ADDR", ":8080"),
        apiURL:      env.GetString("EXTERNAL_URL", "http://localhost:8080"),
        frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
        db: dbConfig{
            addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost/billify?sslmode=disable"),
            maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 10),
            maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),
            maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
        },
        env: env.GetString("ENV", "development"),
        auth: authConfig{
            token: tokenConfig{
                accSecret: env.GetString("ACCESS_SECRET", ""),
                refSecret: env.GetString("REFRESH_SECRET", ""),
                iss:       env.GetString("ISSUER", "billify"),
                accExp:    time.Hour * 24,
                refExp:    time.Hour * 24 * 30,
            },
            oauth: map[string]oauthConfig{
                "google": {
                    clientID:     env.GetString("OAUTH_GOOGLE_CLIENT_ID", ""),
                    clientSecret: env.GetString("OAUTH_GOOGLE_CLIENT_SECRET", ""),
                    authURL:      env.GetString("OAUTH_GOOGLE_AUTH_URL", ""),
                    tokenURL:     env.GetString("OAUTH_GOOGLE_TOKEN_URL", ""),
                    redirectURL:  env.GetString("OAUTH_GOOGLE_REDIRECT_URL", ""),
                    scopes:       strings.Split(env.GetString("OAUTH_GOOGLE_SCOPES", "openid profile email"), " "),
                },
            },
        },
    }

    // Logger
    logger := zap.Must(zap.NewProduction()).Sugar()
    defer logger.Sync()

    db, err := db.New(
        cfg.db.addr,
        cfg.db.maxOpenConns,
        cfg.db.maxIdleConns,
        cfg.db.maxIdleTime,
    )
    if err != nil {
        logger.Fatal(err)
    }

    defer db.Close()
    logger.Info("database connection established")

    // Authenticator
    jwtAuth := auth.NewJWTAuthenticator(
        cfg.auth.token.accSecret,
        cfg.auth.token.refSecret,
        cfg.auth.token.iss,
        cfg.auth.token.iss,
        cfg.auth.token.accExp,
        cfg.auth.token.refExp,
    )

    oauthConfigs := make(map[string]auth.OAuthConfigReader)
    for provider, cfg := range cfg.auth.oauth {
        oauthConfigs[provider] = &cfg
    }
    oauthAuth := auth.NewOAuthAuthenticator(oauthConfigs)

    store := store.NewStorage(db)
    pdf := pdf.NewPDFGenerator()

    app := &application{
        config: cfg,
        store:  store,
        logger: logger,
        token:  jwtAuth,
        oauth:  oauthAuth,
        pdf:    pdf,
    }

	// Download fonts
    if err := downloadFonts(); err != nil {
        log.Fatalf("Failed to download fonts: %v", err)
    }

    // Metrics collected
    expvar.NewString("version").Set(version)
    expvar.Publish("database", expvar.Func(func() any {
        return db.Stats()
    }))
    expvar.Publish("goroutines", expvar.Func(func() any {
        return runtime.NumGoroutine()
    }))

    mux := app.mount()

    log.Fatal(app.run(mux))
}