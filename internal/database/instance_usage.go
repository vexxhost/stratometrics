package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/vexxhost/stratometrics/internal/database/models"
	"github.com/vexxhost/stratometrics/internal/database/types"
	"gorm.io/gorm"
)

type InstanceUsage struct {
	ProjectID *types.BinaryUUID `json:"project_id,omitempty"`
	Type      string            `json:"type,omitempty"`
	State     string            `json:"state,omitempty"`
	Image     *types.BinaryUUID `json:"image,omitempty"`
	Usage     int64             `json:"usage"`
}

func GetInstancesUsageForProject(db *gorm.DB, from, to time.Time, projectID types.BinaryUUID, groupBy []string) ([]InstanceUsage, error) {
	var usages []InstanceUsage

	fields := fmt.Sprintf(
		"project_id, timestamp, state, %s",
		strings.Join(groupBy, ", "),
	)

	eventsWithNextPeriodQuery := db.Model(&models.InstanceEvent{}).Select(
		fmt.Sprintf(`
			%s,
			IF (
				LEAD(timestamp) OVER (PARTITION BY uuid ORDER BY timestamp ASC ROWS BETWEEN 1 FOLLOWING AND 1 FOLLOWING) > ?,
				?,
				IF (
					LEAD(timestamp) OVER (PARTITION BY uuid ORDER BY timestamp ASC ROWS BETWEEN 1 FOLLOWING AND 1 FOLLOWING) IS NULL,
					NOW(),
					LEAD(timestamp) OVER (PARTITION BY uuid ORDER BY timestamp ASC ROWS BETWEEN 1 FOLLOWING AND 1 FOLLOWING)
				)
			) AS period_end
		`, fields),
		to, to,
	)

	if projectID != types.EmptyUUID {
		eventsWithNextPeriodQuery = eventsWithNextPeriodQuery.Where("project_id = ?", projectID)
	}

	eventsQuery := db.Debug().Table("(?) AS events_with_next_period_query", eventsWithNextPeriodQuery).Select(
		fmt.Sprintf(`
			%s,
			SUM(
				TIMESTAMPDIFF(
					SECOND,
					IF (timestamp < ?, ?, timestamp),
					period_end
				)
			) AS `+"`usage`"+`
		`, fields),
		from, from,
	).Where(
		"state NOT IN (?)", []string{"deleted", "soft_deleted"},
	).Where(
		"period_end >= ?", from,
	).Where(
		"timestamp <= ?", to,
	).Group(
		fmt.Sprintf("project_id, %s", strings.Join(groupBy, ", ")),
	)

	ret := eventsQuery.Find(&usages)
	if ret.Error != nil {
		return nil, ret.Error
	}

	return usages, nil
}
