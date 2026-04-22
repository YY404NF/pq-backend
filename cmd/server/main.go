package main

import (
	"log"

	"github.com/YY404NF/pq-backend/internal/app"
	"github.com/YY404NF/pq-backend/internal/config"
)

func main() {
	cfg := config.Load()
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
