CREATE TABLE IF NOT EXISTS measurement 
(
    sensor_id UInt64,
    metric LowCardinality(String),
    measured_at DateTime64(3) CODEC(DoubleDelta, LZ4),
    measured_at_date Date MATERIALIZED toDate(measured_at) CODEC(Delta, LZ4),
    received_at DateTime64(3) CODEC(DoubleDelta, ZSTD(3)),
    value Float64 CODEC(Gorilla, LZ4),
    attributes Map(LowCardinality(String), String)
)
ENGINE = ReplacingMergeTree(received_at)
ORDER BY (metric, sensor_id, measured_at)
PARTITION BY toYYYYMM(measured_at_date);