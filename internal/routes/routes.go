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

func SetupRouter(authHandler *handlers.AuthHandler, steamHandler *handlers.SteamHandler, userHandler *handlers.UserHandler, gameHandler *handlers.GameHandler, reviewHandler *handlers.ReviewHandler, listHandler *handlers.ListHandler, likeHandler *handlers.LikeHandler, notifHandler *handlers.NotificationHandler, adminHandler *handlers.AdminHandler) *gin.Engine {
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
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/facebook", authHandler.FacebookLogin)
		auth.GET("/facebook/callback", authHandler.FacebookCallback)
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
		users.GET("/stats", userHandler.GetStats)
		users.GET("/diary", userHandler.GetDiary)
		users.GET("/watchlist", userHandler.GetWatchlist)
		users.GET("/followers", userHandler.GetFollowers)
		users.GET("/following", userHandler.GetFollowing)

		usersProtected := users.Group("")
		usersProtected.Use(middleware.Authenticate())
		{
			usersProtected.GET("/me", userHandler.Me)
			usersProtected.POST("/follow", userHandler.FollowUser)
		}
	}

	games := v1.Group("/games")
	{
		games.GET("/trending", gameHandler.TrendingGames)
		games.GET("/popular", gameHandler.PopularGames)
		games.GET("/search", gameHandler.SearchGames)
		games.GET("/:id", gameHandler.GetGameDetail)
		games.GET("/:id/likes", likeHandler.GetGameLikes)
		games.GET("/:id/reviews", reviewHandler.GetGameReviews)
		games.GET("/:id/lists", listHandler.GetGameLists)

		gamesProtected := games.Group("")
		gamesProtected.Use(middleware.Authenticate())
		{
			gamesProtected.POST("/rate", gameHandler.RateGame)
		}
	}

	reviews := v1.Group("/reviews")
	{
		reviews.GET("/trending", reviewHandler.TrendingReviews)
		reviews.GET("/:id", reviewHandler.GetReviewDetail)
		reviews.GET("/:id/comments", reviewHandler.GetComments)

		reviewsProtected := reviews.Group("")
		reviewsProtected.Use(middleware.Authenticate())
		{
			reviewsProtected.POST("", reviewHandler.CreateReview)
			reviewsProtected.PUT("/:id", reviewHandler.UpdateReview)
			reviewsProtected.DELETE("/:id", reviewHandler.DeleteReview)
			reviewsProtected.POST("/:id/comments", reviewHandler.AddComment)
		}
	}

	lists := v1.Group("/lists")
	{
		lists.GET("/trending", listHandler.TrendingLists)
		lists.GET("/:id", listHandler.GetListDetail)

		listsProtected := lists.Group("")
		listsProtected.Use(middleware.Authenticate())
		{
			listsProtected.POST("", listHandler.CreateList)
			listsProtected.PUT("/:id", listHandler.UpdateList)
			listsProtected.DELETE("/:id", listHandler.DeleteList)
		}
	}

	likesGroup := v1.Group("")
	likesGroup.Use(middleware.Authenticate())
	{
		likesGroup.POST("/games/:id/like", likeHandler.LikeGame)
		likesGroup.POST("/reviews/:id/like", likeHandler.LikeReview)
		likesGroup.POST("/lists/:id/like", likeHandler.LikeList)
	}

	notifications := v1.Group("/notifications")
	notifications.Use(middleware.Authenticate())
	{
		notifications.GET("", notifHandler.GetNotifications)
		notifications.POST("/:id/read", notifHandler.MarkAsRead)
		notifications.POST("/read-all", notifHandler.MarkAllAsRead)
	}

	admin := v1.Group("/admin")
	admin.Use(middleware.Authenticate(), middleware.RequireAdmin())
	{
		admin.GET("/stats", adminHandler.GetDashboardStats)
		admin.GET("/chart", adminHandler.GetActivityChart)
		admin.GET("/games", adminHandler.GetGames)
		users := admin.Group("/users")
		{
			users.GET("/:id", adminHandler.GetUserDetail)
			users.PATCH("/:id/status", adminHandler.UpdateStatus)
			users.PATCH("/:id/role", adminHandler.UpdateRole)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ép buộc cho phép cổng 5173 để xử lý triệt để lỗi CORS
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
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
