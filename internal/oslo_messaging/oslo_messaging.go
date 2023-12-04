package oslo_messaging

import "encoding/json"

type Message struct {
	Version string `json:"oslo.version"`
	Message string `json:"oslo.message"`
}

func Unmarshal(body []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
