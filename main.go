package main

import (
    "flag"
    "vault/be/database"
    "vault/be/internal/app"
)

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