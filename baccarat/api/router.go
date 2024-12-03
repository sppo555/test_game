package api

import (
	"baccarat/api/handlers"
	"baccarat/api/middleware"
	"baccarat/internal/auth"
	"database/sql"
	"net/http"
)

type Router struct {
	authHandler *handlers.AuthHandler
	userHandler *handlers.UserHandler
	gameHandler *handlers.GameHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(db *sql.DB) *Router {
	jwtService := auth.NewJWTService()
	return &Router{
		authHandler:    handlers.NewAuthHandler(db, jwtService),
		userHandler:    handlers.NewUserHandler(db),
		gameHandler:    handlers.NewGameHandler(db),
		authMiddleware: middleware.NewAuthMiddleware(jwtService),
	}
}

func (r *Router) SetupRoutes() {
	// 公開路由
	http.HandleFunc("/api/register", r.authHandler.Register)
	http.HandleFunc("/api/login", r.authHandler.Login)

	// 需要認證的路由
	http.Handle("/api/user/balance", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetBalance)))
	http.Handle("/api/user/deposit", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.Deposit)))
	http.Handle("/api/user/transactions", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetTransactions)))
	http.Handle("/api/user/bets", r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetBets)))
	http.Handle("/api/game/play", r.authMiddleware.Authenticate(http.HandlerFunc(r.gameHandler.PlayGame)))
	http.Handle("/api/logout", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Logout)))
}
