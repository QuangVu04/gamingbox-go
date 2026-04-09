package routes

import (
	"net/http"
	"vault/be/config"
	"vault/be/internal/handlers"
	"vault/be/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(authHandler *handlers.AuthHandler, steamHandler *handlers.SteamHandler, userHandler *handlers.UserHandler, gameHandler *handlers.GameHandler, reviewHandler *handlers.ReviewHandler, listHandler *handlers.ListHandler, likeHandler *handlers.LikeHandler) *gin.Engine {
	if config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "env": config.App.Env})
	})

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.GET("/steam", steamHandler.LoginHandle)
		auth.GET("/steam/callback", steamHandler.CallbackHandle)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		protected := auth.Group("")
		protected.Use(middleware.Authenticate())
		{
			protected.GET("/me", authHandler.Me)
			protected.POST("/logout", authHandler.Logout)
		}
	}

	users := v1.Group("/users")
	users.Use(middleware.Authenticate())
	{
		users.GET("/me", userHandler.Me)
	}

	games := v1.Group("/games")
	{
		games.GET("/trending", gameHandler.TrendingGames)

		gamesProtected := games.Group("")
		gamesProtected.Use(middleware.Authenticate())
		{
			gamesProtected.POST("/rate", gameHandler.RateGame)
		}
	}

	reviews := v1.Group("/reviews")
	{
		reviews.GET("/trending", reviewHandler.TrendingReviews)
	}

	lists := v1.Group("/lists")
	{
		lists.GET("/trending", listHandler.TrendingLists)
	}

	likesGroup := v1.Group("")
	likesGroup.Use(middleware.Authenticate())
	{
		likesGroup.POST("/games/:game_id/like", likeHandler.LikeGame)
		likesGroup.POST("/reviews/:review_id/like", likeHandler.LikeReview)
		likesGroup.POST("/lists/:list_id/like", likeHandler.LikeList)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", config.App.FrontEndUrl)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
