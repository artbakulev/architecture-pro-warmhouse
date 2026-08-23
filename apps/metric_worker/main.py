import asyncio
import datetime
import json
import logging
import os
import sys
from aiokafka import AIOKafkaConsumer, ConsumerRecord
from dataclasses import asdict, dataclass, field
from typing import AsyncIterator
from asynch import Connection, DictCursor

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
    stream=sys.stdout,
)

logger = logging.getLogger("metric_worker")


@dataclass(frozen=True, kw_only=True, slots=True)
class KafkaConfig:
    topic: str
    bootstrap_servers: str
    group_id: str


@dataclass(frozen=True, kw_only=True, slots=True)
class ClickhouseConfig:
    dsn: str
    measurement_table: str


@dataclass(frozen=True, kw_only=True, slots=True)
class FlusherConfig:
    seconds_to_flush: int = 100
    items_to_flush: int = 3


@dataclass(frozen=True, kw_only=True, slots=True)
class Config:
    kafka: KafkaConfig
    clickhouse: ClickhouseConfig
    flusher: FlusherConfig

@dataclass(slots=True, kw_only=True)
class Measurement:
    sensor_id: int
    metric: str
    measured_at: datetime.datetime
    value: float
    received_at: datetime.datetime = field(default_factory=datetime.datetime.now)

class KafkaConsumer:
    def __init__(self, config: KafkaConfig):
        self._config = config
        self._consumer = AIOKafkaConsumer(
            config.topic,
            bootstrap_servers=config.bootstrap_servers,
            group_id=config.group_id,
            enable_auto_commit=False,
        )

    async def connect(self):
        await self._consumer.start()
        logger.info(f"Connected to kafka topic {self._config.topic}")

    async def disconnect(self):
        await self._consumer.stop()
        logger.info("Disconnected from kafka")

    async def consume(self) -> AsyncIterator[ConsumerRecord[str, bytes]]:
        async for msg in self._consumer:
            if msg is not None:
                yield msg

    async def commit(self) -> None:
        await self._consumer.commit()


class ClickhouseWriter:
    def __init__(self, config: ClickhouseConfig) -> None:
        self._config = config
        self._connection = Connection(
            dsn=config.dsn
        )

    @property
    def insert_query(self):
        return f"""INSERT INTO {self._config.measurement_table}
        (sensor_id,metric,measured_at,received_at,value) VALUES"""

    async def connect(self):
        await self._connection.connect()
        logger.info("Connected to clickhouse")

    async def disconnect(self):
        if not self._connection.closed:
            await self._connection.close()
            logger.info("Disconnected from clickhouse")

    async def write(self, items: list[Measurement]) -> None:
        async with self._connection.cursor(cursor=DictCursor) as cursor:
            await cursor.execute(
                self.insert_query,
                [asdict(item) for item in items],
            )


class FlushManager:
    def __init__(
        self, config: FlusherConfig, consumer: KafkaConsumer, writer: ClickhouseWriter
    ) -> None:
        self._config = config
        self._consumer = consumer
        self._writer = writer
        self._buffer: list[Measurement] | None = None
        self._last_flushed_at: datetime.datetime | None = None
        self._flush_by_time_task: asyncio.Task | None = None


    async def run(self):
        self._flush_by_time_task = asyncio.create_task(self._flush_by_time())
        while True:
            try:
                await self._run()
            except asyncio.CancelledError:
                await self._flush()
                raise
            except Exception as e:
                logger.exception("Exception while handling metrics", exc_info=e)

    
    async def _flush_by_time(self):
        logger.info("Starting inteval flushing")
        self._last_flushed_at = datetime.datetime.now()
        while True:
            try:
                delta_seconds = (datetime.datetime.now() - self._last_flushed_at).seconds
                if delta_seconds >= self._config.seconds_to_flush:
                    logger.info("Flushing by time...")
                    await self._flush()

                sleep_seconds = (self._last_flushed_at + datetime.timedelta(seconds=self._config.seconds_to_flush) - datetime.datetime.now()).total_seconds()
                if sleep_seconds > 0:
                    await asyncio.sleep(sleep_seconds)
                elif self._buffer is None or len(self._buffer) <= 0:
                    await asyncio.sleep(self._config.seconds_to_flush)
            except Exception as e:
                logger.exception("Exception in flushing by time", exc_info=e)


    async def _maybe_flush(self):
        if self._buffer and len(self._buffer) < self._config.items_to_flush:
            return

        await self._flush()

    async def _flush(self):
        if not self._buffer or len(self._buffer) < 1:
            return

        assert self._buffer is not None
        await self._writer.write(self._buffer)
        await self._consumer.commit()
        logger.info(f"Flushed {len(self._buffer)} items")

        self._buffer.clear()
        self._last_flushed_at = datetime.datetime.now()

    async def _run(self):
        async for msg in self._consumer.consume():
            if msg.value is None:
                logger.error("Empty payload for measurement", extra={"sensor_id": msg.key})
                continue
            logger.info(f"Handling new message for {msg.key=}")
            try:
                raw_value = json.loads(msg.value.decode("UTF-8"))
                measurement = Measurement(
                    sensor_id=int(raw_value["sensor_id"]),
                    metric=str(raw_value["metric"]),
                    measured_at=datetime.datetime.fromisoformat(raw_value["measured_at"]),
                    value=float(raw_value["value"]),
                )
            except Exception as e:
                logger.exception("Can not parse payload", exc_info=e)

            if not self._buffer:
                self._buffer = [measurement]
            else:
                self._buffer.append(measurement)
            
            await self._maybe_flush()


async def main():
    logger.info("Starting worker...")
    config = Config(
        kafka=KafkaConfig(
            topic=os.environ.get("METRIC_WORKER_KAFKA_TOPIC", "measurement.result"),
            group_id=os.environ.get("METRIC_WORKER_KAFKA_GROUP_ID", "1"),
            bootstrap_servers=os.environ.get(
                "METRIC_WORKER_KAFKA_BOOTSTRAP_SERVERS", "kafka:9092"
            ),
        ),
        clickhouse=ClickhouseConfig(
            dsn=os.environ.get("CLICKHOUSE_DSN", "clickhouse://smarthome:clickhouse@clickhouse:9000"),
            measurement_table=os.environ.get("CLICKHOUSE_TABLE_NAME", "measurement"),
        ),
        flusher=FlusherConfig(),
    )
    writer = ClickhouseWriter(config=config.clickhouse)
    consumer = KafkaConsumer(config=config.kafka)
    manager = FlushManager(config=config.flusher, consumer=consumer, writer=writer)
    try:
        await writer.connect()
        await consumer.connect()
        await manager.run()
    finally:
        await writer.disconnect()
        await consumer.disconnect()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        exit(0)
