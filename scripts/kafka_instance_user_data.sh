#!/bin/bash
apt-get update -y
apt-get install -y docker.io

systemctl start docker
systemctl enable docker

# Get the private IP of this instance
PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)

# Run Zookeeper
docker run -d --name zookeeper --restart always \
  -p 2181:2181 \
  -e ZOOKEEPER_CLIENT_PORT=2181 \
  -e ZOOKEEPER_TICK_TIME=2000 \
  confluentinc/cp-zookeeper:7.5.0

# Wait for Zookeeper to start
sleep 10

# Run Kafka
docker run -d --name kafka --restart always \
  -p 9092:9092 \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=${PRIVATE_IP}:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://${PRIVATE_IP}:9092 \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
  -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
  confluentinc/cp-kafka:7.5.0