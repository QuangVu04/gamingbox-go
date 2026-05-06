package routes

import (
	"net/http"
	"vault/be/config"
	"vault/be/internal/handlers"
	"vault/be/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Swagger UI Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.GET("/steam", steamHandler.LoginHandle)
		auth.GET("/steam/callback", steamHandler.CallbackHandle)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/verify-code", authHandler.VerifyCode)
		auth.POST("/reset-password", authHandler.ResetPassword)

		protected := auth.Group("")
		protected.Use(middleware.Authenticate())
		{
			protected.GET("/me", authHandler.Me)
			protected.POST("/logout", authHandler.Logout)
		}
	}

	users := v1.Group("/users")
	{
		users.GET("/profile", userHandler.GetProfile)

		usersProtected := users.Group("")
		usersProtected.Use(middleware.Authenticate())
		{
			usersProtected.GET("/me", userHandler.Me)
			usersProtected.GET("/me/following", userHandler.GetFollowing)
			usersProtected.POST("/follow", userHandler.FollowUser)
		}
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
