package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/config"
	userDB "github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/db"
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/endpoint"
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/jwt"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/keys"
)

func main() {
	//ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("ENV loaded...")

	db, err := dbInitializer.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	log.Println("DB client initialized...")

	err = userDB.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("DB found...")

	privKey, err := keys.GetPrivateKey(cfg.PrivateKeySecretPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	pubKey, err := keys.GetPublicKey(cfg.PublicJwtSecretPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Keys loaded...")

	tokenAuth := jwt.InitAuth(privKey, pubKey)
	log.Println("Token auth initialized...")

	r := endpoint.Init(db, tokenAuth)
	log.Println("Endpoint initialized...")

	log.Println("Auth Server running on port: " + cfg.Port + "...")
	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
		return
	}

}
