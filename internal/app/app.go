package app

import (
	"log"
	"vault/be/config"
	"vault/be/database"
	"vault/be/internal/cron"
	"vault/be/internal/handlers"
	"vault/be/internal/repositories"
	"vault/be/internal/routes"
	"vault/be/internal/seeders"
	"vault/be/internal/services"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func New() *App {
	// 1. Load config FIRST
	config.Load()

	// 2. Connect to Database SECOND
	db := database.Connect()

	// 3. Connect to Redis
	err := database.InitRedis()
	if err != nil {
		log.Fatalf("Lỗi khởi tạo Redis: %v", err)
	}
	return &App{DB: db, RDB: database.RDB}
}

func (a *App) Seed() {
	seeders.SeedAdmin(a.DB)
	seeders.SeedRandomData(a.DB)
}

func (a *App) Run() {
	userRepo := repositories.NewUserRepository(a.DB)
	tokenRepo := repositories.NewRefreshTokenRepository(a.DB)
	reviewRepo := repositories.NewReviewRepository(a.DB)
	gameLogRepo := repositories.NewGameLogRepository(a.DB)
	listRepo := repositories.NewListRepository(a.DB)
	activityLogRepo := repositories.NewActivityLogRepository(a.DB)
	ratingRepo := repositories.NewRatingRepository(a.DB)
	gameRepo := repositories.NewGameRepository(a.DB)
	likeRepo := repositories.NewLikeRepository(a.DB)

	authService := services.NewAuthService(userRepo, tokenRepo)
	userService := services.NewUserService(userRepo, reviewRepo, gameLogRepo, listRepo, activityLogRepo, ratingRepo)
	gameService := services.NewGameService(gameRepo, ratingRepo, a.DB, a.RDB)
	reviewService := services.NewReviewService(reviewRepo, a.RDB)
	listService := services.NewListService(listRepo, a.RDB)
	likeService := services.NewLikeService(likeRepo, a.RDB)

	authH := handlers.NewAuthHandler(authService)
	steamH := handlers.NewSteamHandler(authService)
	userH := handlers.NewUserHandler(userService)
	gameH := handlers.NewGameHandler(gameService)
	reviewH := handlers.NewReviewHandler(reviewService)
	listH := handlers.NewListHandler(listService)
	likeH := handlers.NewLikeHandler(likeService)

	cronMgr, err := cron.NewCronManager(a.RDB, gameService, reviewService, listService)
	if err != nil {
		log.Fatal("Không thể tạo Cron Manager:", err)
	}
	if err := cronMgr.Start(); err != nil {
		log.Fatal("Không thể start Cron jobs:", err)
	}
	// 4. Setup Routes
	r := routes.SetupRouter(authH, steamH, userH, gameH, reviewH, listH, likeH)

	log.Printf("Server starting on port %s", config.App.Port)
	r.Run(":" + config.App.Port)
}
