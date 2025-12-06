package main

import (
    //"database/sql"
    "github.com/joho/godotenv"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/streadway/amqp"

    "sms_leopard/controllers"
    dbpkg "sms_leopard/db"
    "sms_leopard/models"
    "sms_leopard/queue"
    workerpkg "sms_leopard/worker"
)

func main(){
    godotenv.Load() 
    dsn := os.Getenv("DSN")
    if dsn=="" { log.Fatal("DSN env required") }
    sqlDB, err := dbpkg.OpenFromDSN(dsn)
    if err!=nil { log.Fatalf("db open: %v", err) }
    if err := dbpkg.Migrate(sqlDB); err!=nil { log.Fatalf("migrate: %v", err) }
    svc := models.NewService(sqlDB)

    amqpURL := os.Getenv("AMQP_URL")
    if amqpURL=="" { log.Fatal("AMQP_URL env required") }
    conn, err := amqp.Dial(amqpURL)
    if err!=nil { log.Fatalf("amqp dial: %v", err) }
    defer conn.Close()

    pub, err := queue.NewPublisher(conn, "smsleopard-exchange")
    if err!=nil { log.Fatalf("publisher: %v", err) }
    consumer, err := queue.NewConsumer(conn, "smsleopard-queue")
    if err!=nil { log.Fatalf("consumer: %v", err) }

    handler := controllers.NewHandler(svc, pub)
    mux := http.NewServeMux()
    mux.HandleFunc("/campaigns", func(w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{ handler.CreateCampaign(w,r); return }
        if r.Method=="GET"{ handler.ListCampaigns(w,r); return }
        w.WriteHeader(http.StatusMethodNotAllowed)
    })
    mux.HandleFunc("/campaigns/send", func(w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{ handler.SendCampaign(w,r); return }
        w.WriteHeader(http.StatusMethodNotAllowed)
    })
    mux.HandleFunc("/preview", func(w http.ResponseWriter, r *http.Request){ if r.Method=="POST"{ handler.Preview(w,r); return }; w.WriteHeader(http.StatusMethodNotAllowed)})

    worker := workerpkg.NewWorker(svc, consumer)
    go func(){
        if err := worker.Start(); err!=nil { log.Printf("worker err: %v", err) }
    }()

    srv := &http.Server{ Addr: ":8080", Handler: mux, ReadTimeout: 10*time.Second, WriteTimeout: 10*time.Second }
    log.Println("listening :8080"); log.Fatal(srv.ListenAndServe())
}
