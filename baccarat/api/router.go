package api

import (
	"baccarat/api/handlers"
	"baccarat/api/middleware"
	"baccarat/internal/auth"
	"database/sql"
	"net/http"
)

type Router struct {
	mux           *http.ServeMux
	authHandler   *handlers.AuthHandler
	userHandler   *handlers.UserHandler
	gameHandler   *handlers.GameHandler
	autoGameHandler *handlers.AutoGameHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(db *sql.DB) *Router {
	jwtService := auth.NewJWTService()
	router := &Router{
		mux:           http.NewServeMux(),
		authHandler:   handlers.NewAuthHandler(db, jwtService),
		userHandler:   handlers.NewUserHandler(db),
		gameHandler:   handlers.NewGameHandler(db),
		autoGameHandler: handlers.NewAutoGameHandler(db),
		authMiddleware: middleware.NewAuthMiddleware(jwtService),
	}
	router.setupRoutes()
	return router
}

func (r *Router) setupRoutes() {
	// 公開路由
	r.mux.HandleFunc("/api/register", r.authHandler.Register)
	r.mux.HandleFunc("/api/login", r.authHandler.Login)

	// 需要認證的路由
	r.mux.Handle("/api/user/balance", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetBalance)))
	r.mux.Handle("/api/user/deposit", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.Deposit)))
	r.mux.Handle("/api/user/transactions", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetTransactions)))
	r.mux.Handle("/api/user/bets", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetBets)))
	r.mux.Handle("/api/game/play", r.authMiddleware.Authenticate(http.HandlerFunc(r.gameHandler.PlayGame)))
	r.mux.Handle("/api/logout", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Logout)))

	// 自動遊戲路由
	r.mux.Handle("/api/auto-game/status", r.authMiddleware.Authenticate(http.HandlerFunc(r.autoGameHandler.GetAutoGameStatus)))
	r.mux.Handle("/api/auto-game/bet", r.authMiddleware.Authenticate(http.HandlerFunc(r.autoGameHandler.PlaceAutoBet)))
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
