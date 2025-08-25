package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectToRabbitMQ(t *testing.T) {
	// Пропускаем тест, если RABBITMQ_URL не установлен
	if os.Getenv("RABBITMQ_URL") == "" {
		t.Skip("Skipping test: RABBITMQ_URL not set")
	}

	tests := []struct {
		name        string
		url         string
		queueName   string
		expectError bool
	}{
		{
			name:        "ValidConnection",
			url:         os.Getenv("RABBITMQ_URL"),
			queueName:   "test_queue",
			expectError: false,
		},
		{
			name:        "InvalidURL",
			url:         "amqp://guest:guest@wrong:5672/",
			queueName:   "test_queue",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, ch, q, err := connectToRabbitMQ(tc.url, tc.queueName)
			if tc.expectError {
				assert.Error(t, err, "expected error for %s", tc.name)
			} else {
				assert.NoError(t, err, "unexpected error for %s", tc.name)
				assert.Equal(t, tc.queueName, q.Name, "queue name mismatch")
				// Проверяем возможность потребления из очереди
				_, err = ch.Consume(q.Name, "", true, false, false, false, nil)
				assert.NoError(t, err, "failed to consume from queue")
				if conn != nil {
					conn.Close()
				}
				if ch != nil {
					ch.Close()
				}
			}
		})
	}
}
