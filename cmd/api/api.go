package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"billify-api/internal/auth"
	"billify-api/internal/env"
	"billify-api/internal/pdf"
	"billify-api/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	token  auth.Authenticator
	oauth  auth.OAuthAuthenticator
	pdf    pdf.PDFGenerator
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	frontendURL string
	auth        authConfig
}

type authConfig struct {
	token tokenConfig
	oauth map[string]oauthConfig
}

type tokenConfig struct {
	accSecret string
	refSecret string
	iss       string
	accExp    time.Duration
	refExp    time.Duration
}

type oauthConfig struct {
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	redirectURL  string
	scopes       []string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (c *oauthConfig) GetClientID() string     { return c.clientID }
func (c *oauthConfig) GetClientSecret() string { return c.clientSecret }
func (c *oauthConfig) GetAuthURL() string      { return c.authURL }
func (c *oauthConfig) GetTokenURL() string     { return c.tokenURL }
func (c *oauthConfig) GetRedirectURL() string  { return c.redirectURL }
func (c *oauthConfig) GetScopes() []string     { return c.scopes }

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGINS", app.config.frontendURL)},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Route("/health", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Get("/", app.healthCheckHandler)
		})
		r.Route("/auth", func(r chi.Router) {
			r.Get("/refresh", app.refreshTokenHandler)
			r.Post("/register", app.registerUserHandler)
			r.Post("/login", app.loginHandler)
			r.Post("/logout", app.logoutHandler)
		})
		r.Route("/oauth", func(r chi.Router) {
			r.Get("/{provider}", app.providerOAuthHandler)
			r.Get("/{provider}/callback", app.callbackOAuthHandler)
		})
		r.Route("/user", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Get("/", app.getUserHandler)
			r.Put("/", app.updateUserHandler)
		})
		r.Route("/business", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Get("/", app.getBusinessesByUserIDHandler)
			r.Post("/dashboard", app.getBusinessesDashboardHandler)
			r.Post("/", app.createBusinessHandler)
			r.Put("/", app.updateBusinessHandler)
			r.Get("/{busID}", app.getBusinessByIDHandler)
		})
		r.Route("/invoices", func(r chi.Router) {
            r.Use(app.AuthMiddleware)
            r.Post("/", app.createInvoiceHandler)
            r.Get("/{id}", app.getInvoiceByIDHandler)
            r.Put("/", app.updateInvoiceHandler)
			r.Put("/{id}/status", app.updateInvoiceStatusHandler)
            r.Delete("/{id}", app.deleteInvoiceHandler)
			r.Get("/{invID}/pdf", app.getInvoiceAsPDFHandler)
            r.Get("/business/{busID}", app.getInvoicesByBusinessIDHandler)
			r.Get("/next-invoice-no/{busID}", app.getNextInvoiceNumberHandler)
        })
		r.Route("/customers", func(r chi.Router) {
		    r.Use(app.AuthMiddleware)
		    r.Get("/business/{busID}", app.getCustomersByBusinessIDHandler)
		    r.Post("/", app.createCustomerHandler)
		    r.Put("/", app.updateCustomerHandler)
		    r.Delete("/{id}", app.deleteCustomerHandler)
		    // r.Get("/{id}", app.getCustomerByIDHandler)
		})
		r.Route("/products", func(r chi.Router) {
		    r.Use(app.AuthMiddleware)
		    r.Post("/", app.createProductHandler)
		    r.Put("/", app.updateProductHandler)
		    r.Delete("/{id}", app.deleteProductHandler)
		    r.Get("/business/{busID}", app.getProductsByBusinessIDHandler)
		    // r.Get("/{id}", app.getProductByIDHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
