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

func SetupRouter(authHandler *handlers.AuthHandler, steamHandler *handlers.SteamHandler, userHandler *handlers.UserHandler, gameHandler *handlers.GameHandler, reviewHandler *handlers.ReviewHandler, listHandler *handlers.ListHandler, likeHandler *handlers.LikeHandler, notifHandler *handlers.NotificationHandler, adminHandler *handlers.AdminHandler, uploadHandler *handlers.UploadHandler) *gin.Engine {
	if config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	// Phân phối file tĩnh trong thư mục uploads
	r.Static("/uploads", "./uploads")

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
		auth.POST("/register/request-otp", authHandler.RequestRegisterOTP)
		auth.POST("/register/verify", authHandler.VerifyRegisterOTP)
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

	upload := v1.Group("/upload")
	upload.Use(middleware.Authenticate())
	{
		upload.POST("/image", uploadHandler.UploadImage)
	}

	users := v1.Group("/users")
	{
		users.GET("/profile", userHandler.GetProfile)
		users.GET("/stats", userHandler.GetStats)
		users.GET("/diary", userHandler.GetDiary)
		users.GET("/watchlist", userHandler.GetWatchlist)
		users.GET("/followers", userHandler.GetFollowers)
		users.GET("/following", userHandler.GetFollowing)
		users.GET("/reviews", userHandler.GetReviews)
		users.GET("/lists", userHandler.GetLists)
		users.GET("/activities", userHandler.GetActivities)

		usersProtected := users.Group("")
		usersProtected.Use(middleware.Authenticate())
		{
			usersProtected.GET("/me", userHandler.Me)
			usersProtected.PUT("/me", userHandler.UpdateProfile)
			usersProtected.POST("/me/email/request-otp", userHandler.RequestEmailChangeOTP)
			usersProtected.PUT("/me/email/verify", userHandler.VerifyEmailChangeOTP)
			usersProtected.POST("/follow", userHandler.FollowUser)
			usersProtected.PUT("/favorite-games", userHandler.UpdateFavoriteGames)
		}
	}

	games := v1.Group("/games")
	{
		games.GET("/trending", gameHandler.TrendingGames)
		games.GET("/popular", gameHandler.PopularGames)
		games.GET("/search", gameHandler.SearchGames)
		games.GET("/steam-search", gameHandler.SearchSteamGames)
		games.GET("/steam-detail", gameHandler.GetSteamAppDetails)
		games.GET("/genres", gameHandler.GetGenres)
		games.GET("/platforms", gameHandler.GetPlatforms)
		games.GET("/studios", gameHandler.SearchStudios)
		games.GET("/steam-studios", gameHandler.SearchSteamStudios)
		games.GET("/:id", middleware.OptionalAuth(), gameHandler.GetGameDetail)
		games.GET("/:id/likes", likeHandler.GetGameLikes)
		games.GET("/:id/reviews", middleware.OptionalAuth(), reviewHandler.GetGameReviews)
		games.GET("/:id/lists", listHandler.GetGameLists)

		gamesProtected := games.Group("")
		gamesProtected.Use(middleware.Authenticate())
		{
			gamesProtected.GET("/:id/state", gameHandler.GetGameUserState)
			gamesProtected.POST("/log", gameHandler.LogGameStatus)
			gamesProtected.POST("/rate", gameHandler.RateGame)
			gamesProtected.POST("", middleware.RequireAdmin(), gameHandler.CreateGame)
			gamesProtected.PUT("/:id", middleware.RequireAdmin(), gameHandler.UpdateGame)
			gamesProtected.DELETE("/:id", middleware.RequireAdmin(), gameHandler.DeleteGame)
			gamesProtected.DELETE("/genres/:name", middleware.RequireAdmin(), gameHandler.DeleteGenre)
			gamesProtected.DELETE("/platforms/:name", middleware.RequireAdmin(), gameHandler.DeletePlatform)
		}
	}

	reviews := v1.Group("/reviews")
	{
		reviews.GET("/trending", middleware.OptionalAuth(), reviewHandler.TrendingReviews)
		reviews.GET("/:id", middleware.OptionalAuth(), reviewHandler.GetReviewDetail)
		reviews.GET("/:id/comments", middleware.OptionalAuth(), reviewHandler.GetComments)

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
		lists.GET("/trending", middleware.OptionalAuth(), listHandler.TrendingLists)
		lists.GET("/:id", middleware.OptionalAuth(), listHandler.GetListDetail)
		lists.GET("/:id/comments", middleware.OptionalAuth(), listHandler.GetComments)

		listsProtected := lists.Group("")
		listsProtected.Use(middleware.Authenticate())
		{
			listsProtected.POST("", listHandler.CreateList)
			listsProtected.PUT("/:id", listHandler.UpdateList)
			listsProtected.DELETE("/:id", listHandler.DeleteList)
			listsProtected.POST("/:id/comments", listHandler.AddComment)
		}
	}

	studios := v1.Group("/studios")
	{
		studios.GET("/:id", gameHandler.GetStudioDetail)
	}

	likesGroup := v1.Group("")
	likesGroup.Use(middleware.Authenticate())
	{
		likesGroup.POST("/games/:id/like", likeHandler.LikeGame)
		likesGroup.POST("/reviews/:id/like", likeHandler.LikeReview)
		likesGroup.POST("/lists/:id/like", likeHandler.LikeList)
		likesGroup.POST("/comments/:id/like", likeHandler.LikeComment)
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
		admin.GET("/activities", adminHandler.GetAdminActivities)
		admin.DELETE("/reviews/:id", adminHandler.DeleteReview)
		admin.DELETE("/lists/:id", adminHandler.DeleteList)
		users := admin.Group("/users")
		{
			users.GET("", adminHandler.GetUsers)
			users.GET("/:id", adminHandler.GetUserDetail)
			users.GET("/:id/overview", adminHandler.GetUserOverview)
			users.GET("/:id/activities", adminHandler.GetUserActivitiesPaginated)
			users.GET("/:id/reviews", adminHandler.GetUserReviewsPaginated)
			users.GET("/:id/lists", adminHandler.GetUserListsPaginated)
			users.GET("/:id/backlog", adminHandler.GetUserBacklogPaginated)
			users.GET("/:id/games", adminHandler.GetUserGamesPaginated)
			users.PATCH("/:id/status", adminHandler.UpdateStatus)
			users.PATCH("/:id/role", adminHandler.UpdateRole)
			users.PUT("/:id", adminHandler.UpdateUser)
			users.DELETE("/:id", adminHandler.DeleteUser)
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
