package types

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type BinaryUUID uuid.UUID

var (
	EmptyUUID = BinaryUUID(uuid.Nil)
)

func ParseUUID(id string) BinaryUUID {
	return BinaryUUID(uuid.MustParse(id))
}

func (b BinaryUUID) String() string {
	return uuid.UUID(b).String()
}

func (b BinaryUUID) MarshalJSON() ([]byte, error) {
	s := uuid.UUID(b)
	str := "\"" + s.String() + "\""
	return []byte(str), nil
}

func (b *BinaryUUID) UnmarshalJSON(by []byte) error {
	s, err := uuid.ParseBytes(by)
	*b = BinaryUUID(s)
	return err
}

func (BinaryUUID) GormDataType() string {
	return "binary(16)"
}

func (b *BinaryUUID) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	data, err := uuid.FromBytes(bytes)
	*b = BinaryUUID(data)
	return err
}

func (b BinaryUUID) Value() (driver.Value, error) {
	return uuid.UUID(b).MarshalBinary()
}
