package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/vexxhost/stratometrics/internal/database/models"
	"github.com/vexxhost/stratometrics/internal/database/types"
	"github.com/vexxhost/stratometrics/internal/notifications"
)

func Open() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.WithError(err).Warn("could not load .env file")
	}

	db, err := gorm.Open(mysql.Open(os.Getenv("MYSQL_DSN")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.InstanceEvent{})

	return db, nil
}

func UpsertInstanceEventFromNotification(db *gorm.DB, message *notifications.NovaMessage) error {
	imageUUID, err := message.GetImageUUID()
	if err != nil {
		return err
	}

	notificationInstanceEvent := &models.InstanceEvent{
		Timestamp: time.Time(message.Timestamp),
		ProjectID: types.BinaryUUID(message.Payload.ProjectID),
		UUID:      types.BinaryUUID(message.Payload.InstanceID),
		Type:      message.Payload.InstanceType,
		State:     models.InstanceState(message.Payload.State),
		Image:     types.BinaryUUID(imageUUID),
	}

	var dbInstanceEvent *models.InstanceEvent
	ret := db.Where("uuid = ?", message.Payload.InstanceID).Order("timestamp DESC").First(&dbInstanceEvent)
	if ret.Error != nil && ret.Error != gorm.ErrRecordNotFound {
		return ret.Error
	}

	if ret.Error == gorm.ErrRecordNotFound {
		// NOTE(mnaser): This is a corner case where we have no record of a
		//               deleted instance, so we need to create a record for it.
		if strings.Contains(message.Payload.State, "deleted") {
			originalState := message.Payload.State

			notificationInstanceEvent.State = "active"
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.CreatedAt)
			ret := db.Create(notificationInstanceEvent)
			if ret.Error != nil {
				return ret.Error
			}

			// Switch things back to deleted
			notificationInstanceEvent.ID = 0
			notificationInstanceEvent.State = models.InstanceState(originalState)
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.DeletedAt)
		} else {
			// NOTE(mnaser): Since there is no record, we can put the timestamp to
			//               the `created_at` field.
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.CreatedAt)
		}

		ret := db.Create(notificationInstanceEvent)
		if ret.Error != nil {
			return ret.Error
		}

		return nil
	}

	if dbInstanceEvent.Equals(notificationInstanceEvent) {
		return nil
	} else {
		// print diff
		fmt.Println(cmp.Diff(dbInstanceEvent, notificationInstanceEvent))
	}

	if err := db.Create(notificationInstanceEvent).Error; err != nil {
		return err
	}

	return nil
}
