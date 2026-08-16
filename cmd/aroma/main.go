package main

import (
	"log"
	"net/http"
	"os"

	"aroma-maintenance/internal/application"
	"aroma-maintenance/internal/interfaces/httpapi"
)

func main() {
	service := application.NewMaintenanceService()
	handler := httpapi.NewHandler(service)
	addr := ":8080"
	if configured := os.Getenv("AROMA_ADDR"); configured != "" {
		addr = configured
	}
	log.Printf("aroma maintenance desk listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
