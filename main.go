package main

import (
    "flag"
    "vault/be/database"
    "vault/be/internal/app"
)

func main() {
    migrate := flag.Bool("migrate", false, "Chạy database migration")
    flag.Parse()

    application := app.New()

    if *migrate {
        database.RunMigrations(application.DB)
        return
    }

    application.Run()
}