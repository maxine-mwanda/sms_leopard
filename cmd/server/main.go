package main

import (
	//"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"

	"sms_leopard/controllers"
	dbpkg "sms_leopard/db"
	"sms_leopard/models"
	"sms_leopard/queue"
	workerpkg "sms_leopard/worker"
)

func main() {
	// Load env
	godotenv.Load()

	// Database setup
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN env required")
	}
	sqlDB, err := dbpkg.OpenFromDSN(dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := dbpkg.Migrate(sqlDB); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	svc := models.NewService(sqlDB)

	// AMQP setup
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		log.Fatal("AMQP_URL env required")
	}
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("amqp dial: %v", err)
	}
	defer conn.Close()

	pub, err := queue.NewPublisher(conn, "smsleopard-exchange")
	if err != nil {
		log.Fatalf("publisher: %v", err)
	}
	consumer, err := queue.NewConsumer(conn, "smsleopard-queue")
	if err != nil {
		log.Fatalf("consumer: %v", err)
	}

	// Handlers
	handler := controllers.NewHandler(svc, pub)
	r := mux.NewRouter()

	// Campaign endpoints
	r.HandleFunc("/campaigns", handler.CreateCampaign).Methods("POST")
	r.HandleFunc("/campaigns", handler.ListCampaigns).Methods("GET")
	r.HandleFunc("/campaigns/{id:[0-9]+}", handler.GetCampaignDetails).Methods("GET")
	r.HandleFunc("/campaigns/{id:[0-9]+}/send", handler.SendCampaign).Methods("POST")
	r.HandleFunc("/campaigns/{id:[0-9]+}/personalized-preview", handler.Preview).Methods("POST")

	// Stats & health
	r.HandleFunc("/stats", handler.Stats).Methods("GET")
	r.HandleFunc("/health", handler.Health).Methods("GET")

	// Start worker
	worker := workerpkg.NewWorker(svc, consumer)
	go func() {
		if err := worker.Start(); err != nil {
			log.Printf("worker err: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("listening :8080")
	log.Fatal(srv.ListenAndServe())

}
