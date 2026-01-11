package queue

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"meeting-system/shared/config"
)

// buildKafkaTransport creates a Transport with optional TLS/SASL.
func buildKafkaTransport(cfg config.KafkaConfig) *kafka.Transport {
	transport := &kafka.Transport{
		IdleTimeout: 30 * time.Second,
	}

	if cfg.TLS.Enabled {
		transport.TLS = &tls.Config{InsecureSkipVerify: cfg.TLS.InsecureSkipVerify}
	}

	if cfg.SASL.Enabled {
		switch strings.ToUpper(cfg.SASL.Mechanism) {
		case "", "PLAIN":
			transport.SASL = plain.Mechanism{
				Username: cfg.SASL.Username,
				Password: cfg.SASL.Password,
			}
		}
	}

	return transport
}

// buildKafkaDialer builds a Dialer that mirrors the transport auth settings.
func buildKafkaDialer(cfg config.KafkaConfig) *kafka.Dialer {
	transport := buildKafkaTransport(cfg)
	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           transport.TLS,
		SASLMechanism: transport.SASL,
	}
}
