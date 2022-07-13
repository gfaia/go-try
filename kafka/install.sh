#!/bin/bash
KAFKA_VERSION="kafka_2.13-3.2.0"
rm -rf kafka
rm ${KAFKA_VERSION}.tgz

curl -o ${KAFKA_VERSION}.tgz https://dlcdn.apache.org/kafka/3.2.0/${KAFKA_VERSION}.tgz
tar -xzf ${KAFKA_VERSION}.tgz
mv ${KAFKA_VERSION} kafka

