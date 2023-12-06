package consumers

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vexxhost/stratometrics/internal/clickhousedb"
	"github.com/vexxhost/stratometrics/internal/notifications"
	"github.com/vexxhost/stratometrics/internal/oslo_messaging"
	"github.com/wagslane/go-rabbitmq"
)

func NewNovaConsumer(db *clickhousedb.Database, conn *rabbitmq.Conn) (*rabbitmq.Consumer, error) {
	return rabbitmq.NewConsumer(
		conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			osloMessage, err := oslo_messaging.Unmarshal(d.Body)
			if err != nil {
				return rabbitmq.NackRequeue
			}

			message, err := notifications.UnmarshalNovaMessage([]byte(osloMessage.Message))
			if err != nil {
				log.Println("failed to parse message:", err, osloMessage)
				return rabbitmq.NackRequeue
			}

			if !strings.HasPrefix(message.EventType, "compute.instance") {
				return rabbitmq.Ack
			}

			err = db.UpsertInstanceEventFromNotification(context.TODO(), message)
			if err != nil {
				log.Println("failed to upsert instance event:", err, message)
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		"stratometrics",
		rabbitmq.WithConsumerOptionsExchangeName("nova"),
		rabbitmq.WithConsumerOptionsRoutingKey("notifications.info"),
		rabbitmq.WithConsumerOptionsQueueDurable,
		// TODO: make sure we pick up one-by-one so we can run multiple replicas
		// TODO: make sure records dont get lost with queue config
	)
}
