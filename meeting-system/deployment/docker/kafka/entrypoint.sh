#!/usr/bin/env bash
set -euo pipefail

# 默认为单节点 KRaft 模式
KAFKA_HOME=${KAFKA_HOME:-/opt/kafka}
LOG_DIR=${KAFKA_LOG_DIR:-/var/lib/kafka/data}
BROKER_ID=${KAFKA_BROKER_ID:-1}
LISTENERS=${KAFKA_LISTENERS:-PLAINTEXT://:9092}
ADVERTISED_LISTENERS=${KAFKA_ADVERTISED_LISTENERS:-PLAINTEXT://localhost:9092}
CONFIG_FILE=/etc/kafka/server.properties

mkdir -p "${LOG_DIR}"

cat > ${CONFIG_FILE} <<EOF_CONF
process.roles=broker,controller
node.id=${BROKER_ID}
controller.quorum.voters=${BROKER_ID}@localhost:9093
listeners=${LISTENERS},CONTROLLER://:9093
advertised.listeners=${ADVERTISED_LISTENERS}
listener.security.protocol.map=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
controller.listener.names=CONTROLLER
num.network.threads=3
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
log.dirs=${LOG_DIR}
num.partitions=1
num.recovery.threads.per.data.dir=1
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000
zookeeper.connect=
EOF_CONF

if [ ! -f "${LOG_DIR}/meta.properties" ]; then
  echo "Formatting KRaft metadata in ${LOG_DIR}..."
  CLUSTER_ID=${KAFKA_CLUSTER_ID:-$(${KAFKA_HOME}/bin/kafka-storage.sh random-uuid)}
  ${KAFKA_HOME}/bin/kafka-storage.sh format -t "${CLUSTER_ID}" -c "${CONFIG_FILE}"
fi

exec ${KAFKA_HOME}/bin/kafka-server-start.sh ${CONFIG_FILE}
