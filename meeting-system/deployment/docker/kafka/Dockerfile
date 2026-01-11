FROM eclipse-temurin:11-jre

ARG KAFKA_VERSION=3.8.0
ARG SCALA_VERSION=2.13
ENV KAFKA_HOME=/opt/kafka \
    KAFKA_LOG_DIR=/var/lib/kafka/data \
    KAFKA_CLUSTER_ID=CiZkafkaClusterId000000000000001 \
    KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends curl bash ca-certificates netcat-openbsd; \
    rm -rf /var/lib/apt/lists/*; \
    mkdir -p ${KAFKA_HOME} ${KAFKA_LOG_DIR} /etc/kafka /usr/local/bin;

RUN set -eux; \
    curl -fsSL "https://mirrors.aliyun.com/apache/kafka/${KAFKA_VERSION}/kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz" -o /tmp/kafka.tgz; \
    tar -xzf /tmp/kafka.tgz -C /opt; \
    mv /opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION}/* ${KAFKA_HOME}/; \
    rm -rf /tmp/kafka.tgz /opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION};

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 9092
VOLUME ["/var/lib/kafka/data"]
ENTRYPOINT ["/entrypoint.sh"]
