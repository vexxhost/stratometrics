package models

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vexxhost/stratometrics/internal/database/types"
)

type InstanceState string

const (
	InstanceStateActive           InstanceState = "active"
	InstanceStateBuilding         InstanceState = "building"
	InstanceStatePaused           InstanceState = "paused"
	InstanceStateSuspended        InstanceState = "suspended"
	InstanceStateStopped          InstanceState = "stopped"
	InstanceStateRescued          InstanceState = "rescued"
	InstanceStateResized          InstanceState = "resized"
	InstanceStateSoftDeleted      InstanceState = "soft_deleted"
	InstanceStateDeleted          InstanceState = "deleted"
	InstanceStateError            InstanceState = "error"
	InstanceStateShelved          InstanceState = "shelved"
	InstanceStateShelvedOffloaded InstanceState = "shelved_offloaded"
)

type InstanceEvent struct {
	ID        uint             `gorm:"primaryKey"`
	Timestamp time.Time        `gorm:"index"`
	ProjectID types.BinaryUUID `gorm:"index"`
	UUID      types.BinaryUUID
	Type      string        `gorm:"type:varchar(255)"`
	State     InstanceState `gorm:"type:enum('active', 'building', 'paused', 'suspended', 'stopped', 'rescued', 'resized', 'soft_deleted', 'deleted', 'error', 'shelved', 'shelved_offloaded');index"`
	Image     types.BinaryUUID
}

var (
	cmpIgnoreFields = cmpopts.IgnoreFields(
		InstanceEvent{},
		"ID",
		"Timestamp",
	)
)

func (ie *InstanceEvent) Equals(other *InstanceEvent) bool {
	return cmp.Equal(
		ie,
		other,
		cmpIgnoreFields,
	)
}

func (ie *InstanceEvent) Diff(other *InstanceEvent) string {
	return cmp.Diff(
		ie,
		other,
		cmpIgnoreFields,
	)
}
