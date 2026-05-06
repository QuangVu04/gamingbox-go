package main

import (
    "flag"
    "vault/be/database"
    "vault/be/internal/app"
    _ "vault/be/docs" // Swagger docs
)

// @title           Gaming Box API
// @version         1.0
// @description     Hệ thống API cho Gaming Box.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
    migrate := flag.Bool("migrate", false, "Chạy database migration")
    seed := flag.Bool("seed", false, "Chạy database seeder")
    flag.Parse()

    application := app.New()

    if *migrate {
        database.RunMigrations(application.DB)
        return
    }

    if *seed {
        application.Seed()
        return
    }

    application.Run()
}