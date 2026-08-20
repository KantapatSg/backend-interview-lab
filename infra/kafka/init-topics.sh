#!/usr/bin/env bash
set -euo pipefail

broker="${KAFKA_BROKER:-kafka:9092}"
for topic in jobs.events.v1 jobs.retry.v1 jobs.dlq.v1; do
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server "$broker" --create --if-not-exists \
    --topic "$topic" --partitions 3 --replication-factor 1
done
