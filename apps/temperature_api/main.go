package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type TemperatureHandler struct {}

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func (t TemperatureHandler) GetTemperature(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	log.Printf("Handling location=%s", location)
	value := randomFloat(0, 40)

	resp := TemperatureResponse{
		Value: value,
		Unit: "C",
		Timestamp: time.Now(),
		Location: "Test location",
		Status: "healthy",
		SensorID: strconv.Itoa(randomInt(10000, 99999)),
		SensorType: "temperature",
		Description: "Test description",
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	temperatureHandler := TemperatureHandler{}
	r.Get("/temperature", temperatureHandler.GetTemperature)
	http.ListenAndServe(":8081", r)
}
