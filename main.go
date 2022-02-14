package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"testtask/controller"
	"testtask/model"
	"testtask/service"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USERNAME"), // "testtask"
		os.Getenv("DB_PASSWORD"), // "testtask"
		os.Getenv("DB_HOST"),     // "127.0.0.1"
		os.Getenv("DB_DATABASE"), // "testtask"
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&model.Price{})

	if err != nil {
		panic(err)
	}

	flag.Parse()
	priceManager := service.CreatePriceService(&model.PriceRepository{DB: db})

	pool := service.NewPool()
	go pool.Run()

	http.HandleFunc("/client1", func(w http.ResponseWriter, r *http.Request) {
		controller.HandleHome("ws_client_1.html", w, r)
	})
	http.HandleFunc("/client2", func(w http.ResponseWriter, r *http.Request) {
		controller.HandleHome("ws_client_2.html", w, r)
	})
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		controller.HandleApi(priceManager, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		controller.HandleWs(pool, priceManager, w, r)
	})

	err = http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
