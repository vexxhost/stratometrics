package consumers

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vexxhost/stratometrics/internal/database"
	"github.com/vexxhost/stratometrics/internal/notifications"
	"github.com/vexxhost/stratometrics/internal/oslo_messaging"
	"github.com/wagslane/go-rabbitmq"
	"gorm.io/gorm"
)

func NewNovaConsumer(db *gorm.DB, conn *rabbitmq.Conn) (*rabbitmq.Consumer, error) {
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

			err = database.UpsertInstanceEventFromNotification(db, message)
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
	)
}
