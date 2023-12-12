package clickhousedb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/vexxhost/stratometrics/internal/notifications"
)

type InstanceEvent struct {
	Timestamp time.Time `ch:"timestamp" json:"timestamp,omitempty"`
	ProjectID uuid.UUID `ch:"project_id" json:"project_id,omitempty"`
	UUID      uuid.UUID `ch:"uuid" json:"uuid,omitempty"`
	Type      string    `ch:"type" json:"type,omitempty"`
	State     string    `ch:"state" json:"state,omitempty"`
	Image     uuid.UUID `ch:"image" json:"image,omitempty"`
}

func (ie *InstanceEvent) Equals(other *InstanceEvent) bool {
	return cmp.Equal(
		ie,
		other,
		cmpopts.IgnoreFields(
			InstanceEvent{},
			"Timestamp",
		),
	)
}

type InstanceUsage struct {
	ProjectID *uuid.UUID `ch:"project_id" json:"project_id,omitempty"`
	Type      string     `ch:"type" json:"type,omitempty"`
	State     string     `ch:"state" json:"state,omitempty"`
	Image     *uuid.UUID `ch:"image" json:"image,omitempty"`
	Usage     int64      `ch:"usage" json:"usage"`
}

func (d *Database) GetInstancesUsageForProject(ctx context.Context, from, to time.Time, projectID string, groupBy []string) ([]InstanceUsage, error) {
	groupByString := strings.Join(groupBy, ", ")

	whereString := ""
	if projectID != "" {
		whereString = fmt.Sprintf(`
		WHERE
			project_id = '%s'
		`, projectID)
	}

	query := fmt.Sprintf(`
	WITH
		toDateTime('%d') AS start_time,
		toDateTime('%d') AS end_range

	SELECT
		%s,
		sum(period_duration) AS usage
	FROM (
		SELECT
			*,
			multiIf(
				timestamp < start_time, start_time,
				timestamp
			) AS period_start,
			multiIf(
				next_period_start > end_range, end_range,
				next_period_start = toDateTime(0), now(),
				next_period_start
			) AS period_end,
			toUnixTimestamp(period_end) - toUnixTimestamp(period_start) AS period_duration
		FROM (
			SELECT
				*,
				any(timestamp) OVER (PARTITION BY uuid ORDER BY timestamp ASC ROWS BETWEEN 1 FOLLOWING AND 1 FOLLOWING) AS next_period_start
			FROM
				instance_events
		)
		WHERE
			state NOT IN ('deleted', 'soft_deleted') AND
			period_end >= start_time AND
			period_start <= end_range
	)
	%s
	GROUP BY
		project_id,
		%s
	`, from.Unix(), to.Unix(), groupByString, whereString, groupByString)

	var result []InstanceUsage
	if err := d.Connection.Select(ctx, &result, query); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) GetLatestInstanceEvent(ctx context.Context, instanceID uuid.UUID) (*InstanceEvent, error) {
	var result []InstanceEvent
	if err := d.Connection.Select(ctx, &result, "SELECT * FROM instance_events WHERE uuid = ? ORDER BY timestamp DESC LIMIT 1", instanceID); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result[0], nil
}

func (d *Database) InsertInstanceEvent(ctx context.Context, instanceEvent *InstanceEvent) error {
	batch, err := d.Connection.PrepareBatch(ctx, "INSERT INTO instance_events")
	if err != nil {
		return err
	}

	err = batch.AppendStruct(instanceEvent)
	if err != nil {
		return err
	}

	if err := batch.Send(); err != nil {
		return err
	}

	return nil
}

func (d *Database) UpsertInstanceEventFromNotification(ctx context.Context, message *notifications.NovaMessage) error {
	imageUUID, err := message.GetImageUUID()
	if err != nil {
		return err
	}

	notificationInstanceEvent := &InstanceEvent{
		Timestamp: time.Time(message.Timestamp),
		ProjectID: message.Payload.ProjectID,
		UUID:      message.Payload.InstanceID,
		Type:      message.Payload.InstanceType,
		State:     message.Payload.State,
		Image:     imageUUID,
	}

	dbInstanceEvent, err := d.GetLatestInstanceEvent(ctx, message.Payload.InstanceID)
	if err != nil {
		return err
	}

	if dbInstanceEvent == nil {
		// NOTE(mnaser): This is a corner case where we have no record of a
		//               deleted instance, so we need to create a record for it.
		if strings.Contains(message.Payload.State, "deleted") {
			originalState := message.Payload.State

			notificationInstanceEvent.State = "active"
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.CreatedAt)
			if err = d.InsertInstanceEvent(ctx, notificationInstanceEvent); err != nil {
				return err
			}

			// Switch things back to deleted
			notificationInstanceEvent.State = originalState
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.DeletedAt)
		} else {
			// NOTE(mnaser): Since there is no record, we can put the timestamp to
			//               the `created_at` field.
			notificationInstanceEvent.Timestamp = time.Time(message.Payload.CreatedAt)
		}

		if err = d.InsertInstanceEvent(ctx, notificationInstanceEvent); err != nil {
			return err
		}

		return nil
	}

	if dbInstanceEvent.Equals(notificationInstanceEvent) {
		return nil
	} else {
		// print diff
		fmt.Println(cmp.Diff(dbInstanceEvent, notificationInstanceEvent))
	}

	if err = d.InsertInstanceEvent(ctx, notificationInstanceEvent); err != nil {
		return err
	}

	return nil
}
