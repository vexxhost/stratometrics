package notifications

import (
	"encoding/json"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
)

type NotificationTimestamp time.Time

func (nt *NotificationTimestamp) UnmarshalJSON(b []byte) error {
	ts := string(b)
	if ts == "" || ts == "\"\"" {
		*nt = NotificationTimestamp(time.Time{})
		return nil
	}

	formats := []string{
		"\"2006-01-02 15:04:05.000000\"",
		"\"2006-01-02T15:04:05.000000\"",
		"\"2006-01-02 15:04:05-07:00\"",
	}

	var parsedTime time.Time
	var err error
	for _, format := range formats {
		parsedTime, err = time.Parse(format, ts)
		if err == nil {
			*nt = NotificationTimestamp(parsedTime)
			return nil
		}
	}

	return err
}

type NovaMessage struct {
	MessageID   string                `json:"message_id"`
	PublisherID string                `json:"publisher_id"`
	EventType   string                `json:"event_type"`
	Priority    string                `json:"priority"`
	Timestamp   NotificationTimestamp `json:"timestamp"`
	Payload     NovaMessagePayload    `json:"payload"`
}

type NovaMessagePayload struct {
	ProjectID        uuid.UUID             `json:"tenant_id"`
	InstanceID       uuid.UUID             `json:"instance_id"`
	InstanceType     string                `json:"instance_type"`
	CreatedAt        NotificationTimestamp `json:"created_at"`
	DeletedAt        NotificationTimestamp `json:"deleted_at"`
	ImageRefURL      string                `json:"image_ref_url"`
	State            string                `json:"state"`
	StateDescription string                `json:"state_description"`
}

func UnmarshalNovaMessage(body []byte) (*NovaMessage, error) {
	var notification NovaMessage
	if err := json.Unmarshal(body, &notification); err != nil {
		return nil, err
	}

	return &notification, nil
}

func (n *NovaMessage) GetImageUUID() (uuid.UUID, error) {
	u, err := url.Parse(n.Payload.ImageRefURL)
	if err != nil {
		return uuid.Nil, err
	}

	imageRef := path.Base(u.Path)
	imageUUID, err := uuid.Parse(imageRef)
	if err != nil {
		return uuid.Nil, nil
	}

	return imageUUID, nil
}
