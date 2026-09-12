package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"kabadiconnect/backend/internal/auth"
	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/database"
	"kabadiconnect/backend/internal/handovers"
	"kabadiconnect/backend/internal/listings"
	mw "kabadiconnect/backend/internal/middleware"
	"kabadiconnect/backend/internal/offers"
	"kabadiconnect/backend/internal/payments"
	"kabadiconnect/backend/internal/prices"
	"kabadiconnect/backend/internal/transactions"
	"kabadiconnect/backend/internal/users"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	client, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(cfg.MongoDatabase)

	handler, err := newHandler(cfg, db)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("KabadiConnect API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	_ = srv.Shutdown(shutdown)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func newHandler(cfg config.Config, db *mongo.Database) (http.Handler, error) {
	userRepo := users.NewRepository(db, cfg.Environment == "development")
	sessionRepo := auth.NewSessionRepository(db)
	lotRepo := listings.NewRepository(db, userRepo)
	offerRepo := offers.NewRepository(db)
	priceRepo := prices.NewRepository(db)
	transactionRepo := transactions.NewRepository(db)
	handoverRepo := handovers.NewRepository(db)
	paymentRepo := payments.NewRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, fn := range []func(context.Context) error{
		userRepo.EnsureIndexes, sessionRepo.EnsureIndexes, lotRepo.EnsureIndexes, offerRepo.EnsureIndexes,
		priceRepo.EnsureIndexes, transactionRepo.EnsureIndexes, handoverRepo.EnsureIndexes, paymentRepo.EnsureIndexes,
	} {
		if err := fn(ctx); err != nil {
			return nil, fmt.Errorf("index setup: %w", err)
		}
	}

	authService := auth.NewService(cfg, userRepo, sessionRepo)
	authHandler := auth.NewHandler(authService, cfg)
	userHandler := users.NewHandler(userRepo)
	lotHandler := listings.NewHandler(lotRepo)
	offerService := offers.NewService(offerRepo, lotRepo, userRepo, cfg.Environment == "development")
	offerHandler := offers.NewHandler(offerService, offerRepo)
	priceHandler := prices.NewHandler(priceRepo)
	transactionHandler := transactions.NewHandler(transactionRepo)
	handoverHandler := handovers.NewHandler(handoverRepo)
	paymentHandler := payments.NewHandler(paymentRepo)

	mux := http.NewServeMux()
	authLimiter := mw.NewFixedWindowLimiter(10, time.Minute)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("POST /api/v1/auth/register", authLimiter.Handler(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /api/v1/auth/login", authLimiter.Handler(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/me", userHandler.Me)
	protected.HandleFunc("PATCH /api/v1/me", userHandler.PatchMe)

	// Canonical shared e-waste contracts.
	protected.HandleFunc("GET /api/v1/lots", lotHandler.List)
	protected.HandleFunc("GET /api/v1/lots/{id}", lotHandler.Detail)
	protected.HandleFunc("POST /api/v1/lots/{id}/offers", offerHandler.Create)
	protected.HandleFunc("GET /api/v1/offers", offerHandler.List)
	protected.HandleFunc("PATCH /api/v1/offers/{id}/cancel", offerHandler.Cancel)
	protected.HandleFunc("GET /api/v1/prices", priceHandler.Board)
	protected.HandleFunc("GET /api/v1/prices/history", priceHandler.History)
	protected.HandleFunc("GET /api/v1/transactions", transactionHandler.List)
	protected.HandleFunc("GET /api/v1/transactions/{id}", transactionHandler.Detail)
	protected.HandleFunc("GET /api/v1/handovers", handoverHandler.List)
	protected.HandleFunc("GET /api/v1/handovers/{id}", handoverHandler.Detail)
	protected.HandleFunc("POST /api/v1/handovers/{id}/confirm", handoverHandler.Confirm)
	protected.HandleFunc("GET /api/v1/payments", paymentHandler.List)

	// V1 aliases preserved while recycler clients migrate.
	protected.HandleFunc("GET /api/v1/price-board", priceHandler.Board)
	protected.HandleFunc("GET /api/v1/listings", lotHandler.List)
	protected.HandleFunc("GET /api/v1/listings/{id}", lotHandler.Detail)
	protected.HandleFunc("POST /api/v1/listings/{id}/offers", offerHandler.Create)

	mux.Handle("/api/v1/", mw.Auth(authService, mw.RecyclerOnly(protected)))
	handler := securityHeaders(cors.New(cors.Options{AllowedOrigins: []string{cfg.FrontendOrigin}, AllowedMethods: []string{"GET", "POST", "PATCH", "OPTIONS"}, AllowedHeaders: []string{"Authorization", "Content-Type"}, AllowCredentials: true}).Handler(mw.CheckOrigin(cfg.FrontendOrigin, mux)))
	return handler, nil
}
