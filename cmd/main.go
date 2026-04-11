// Package main is the entry point for the subscription service API.
//
// @title           Subscription Service API
// @version         1.0
// @description     A subscription management service supporting sport courses, plans, and vouchers.
//
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the token value.
package main

import (
	"log"
	"os"

	"github.com/emadhejazian/subscription_service/internal/config"
	"github.com/emadhejazian/subscription_service/internal/database"
	deliveryhttp "github.com/emadhejazian/subscription_service/internal/delivery/http"

	_ "github.com/emadhejazian/subscription_service/docs"
)

func main() {
	cfg := config.Load()

	// reset drops all tables before AutoMigrate, so it must bypass Connect.
	if len(os.Args) > 1 && os.Args[1] == "reset" {
		rawDB, err := database.ConnectRaw(cfg)
		if err != nil {
			log.Fatalf("failed to connect for reset: %v", err)
		}
		database.Reset(rawDB)
		log.Println("reset and seed complete")
		return
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "seed" {
		database.Seed(db)
		log.Println("seeding complete")
		return
	}

	r := deliveryhttp.SetupRouter(db)

	addr := ":" + cfg.AppPort
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
