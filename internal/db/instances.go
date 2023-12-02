package db

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type InstanceEvent struct {
	ProjectID string `ch:"project_id" json:"project_id,omitempty"`
	Type      string `ch:"type" json:"type,omitempty"`
	State     string `ch:"state" json:"state,omitempty"`
	Image     string `ch:"image" json:"image,omitempty"`
	Usage     int64  `ch:"usage" json:"usage"`
}

func GetInstancesUsageForProject(ctx context.Context, from, to time.Time, projectID string, groupBy []string) ([]InstanceEvent, error) {
	groupByString := strings.Join(groupBy, ", ")

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
				next_period_start = toDateTime(0), end_range,
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
			event_type != 'DELETED' AND
			period_end >= start_time AND
			period_start <= end_range
	)
	WHERE
		project_id = '%s'
	GROUP BY
		project_id,
		%s
	`, from.Unix(), to.Unix(), groupByString, projectID, groupByString)

	var result []InstanceEvent
	if err := Connection.Select(ctx, &result, query); err != nil {
		return nil, err
	}

	return result, nil
}
