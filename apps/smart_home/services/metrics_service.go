package services

import (
	"context"
	"smarthome/timeseries"
	"time"
)

type Metric struct {
	SensorId       uint64    `ch:"sensor_id" json:"sensor_id"`
	Metric         string    `ch:"metric" json:"metric"`
	MeasuredAt     time.Time `ch:"measured_at" json:"measured_at"`
	MeasuredAtDate time.Time `ch:"measured_at_date" json:"measured_at_date"`
	ReceivedAt     time.Time `ch:"received_at"`
	Value          float64   `ch:"value" json:"value"`
}

type MetricsService struct {
	Clickhouse *timeseries.Clickhouse
}

func NewMetricsService(clickhouse *timeseries.Clickhouse) *MetricsService {
	return &MetricsService{
		Clickhouse: clickhouse,
	}
}

func (s *MetricsService) GetMetricsById(sensorID uint64) ([]Metric, error) {
	query := "SELECT sensor_id, metric, measured_at, measured_at_date, received_at, value FROM measurement WHERE sensor_id = ? ORDER BY measured_at DESC LIMIT 100;"
	ctx := context.Background()
	metrics, err := timeseries.Query[Metric](
		ctx,
		s.Clickhouse,
		query,
		sensorID,
	)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
