package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/vexxhost/stratometrics/internal/database"
	"github.com/vexxhost/stratometrics/internal/database/models"
	"github.com/vexxhost/stratometrics/internal/database/types"
	"gorm.io/gorm"
)

const (
	numVMs      = 1000000 // Adjust this to control total VMs
	numTypes    = 100
	dateFormat  = "2006-01-02 15:04:05"
	numImages   = 100  // Number of static image UUIDs
	numProjects = 5000 // Number of static project UUIDs
)

var (
	// Removed "BUILD" from possible states for updates
	states    = []string{"active", "stopped", "suspended", "error"}
	startDate = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
)

func main() {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate static image UUIDs
	images := make([]uuid.UUID, numImages)
	for i := range images {
		images[i] = uuid.New()
	}

	// Generate static project UUIDs
	projects := make([]uuid.UUID, numProjects)
	for i := range projects {
		projects[i] = uuid.New()
	}

	db, err := database.Open()
	if err != nil {
		panic(err)
	}

	for i := 0; i < numVMs; i++ {
		projectID := projects[randGen.Intn(numProjects)] // Randomly pick a project UUID
		vmID := uuid.New()
		vmType := fmt.Sprintf("type-%d", randGen.Intn(numTypes)+1)
		initialStates := []string{"active", "building"}
		state := initialStates[randGen.Intn(len(initialStates))] // Randomly pick initial state
		image := images[randGen.Intn(numImages)]                 // Randomly pick an image UUID
		createTime := randomDateAfter(startDate, randGen)

		writeEvent(db, createTime, projectID, vmID, vmType, state, image)

		// If initial state is "BUILD", ensure a follow-up "UPDATE" to "ACTIVE" or "ERROR"
		lastUpdateTime := createTime
		if state == "BUILD" {
			lastUpdateTime = randomDateAfter(lastUpdateTime, randGen)
			updatedState := "ACTIVE"     // Default to "ACTIVE"
			if randGen.Float32() < 0.1 { // 10% chance to go to "ERROR"
				updatedState = "ERROR"
			}
			writeEvent(db, lastUpdateTime, projectID, vmID, vmType, updatedState, image)
			state = updatedState // Update the state to reflect the change
		}

		for randGen.Float32() < 0.2 { // Simulate random updates
			lastUpdateTime = randomDateAfter(lastUpdateTime, randGen)

			// Store the original values to detect changes
			originalType := vmType
			originalState := state
			originalImage := image

			updateType := randGen.Intn(3) // Randomly choose which attribute to update
			switch updateType {
			case 0:
				for originalType == vmType {
					vmType = fmt.Sprintf("type-%d", randGen.Intn(numTypes)+1) // Update type
				}
			case 1:
				for originalState == state {
					state = states[randGen.Intn(len(states))] // Update state ensuring it's never "BUILD"
				}
			case 2:
				for originalImage == image {
					image = images[randGen.Intn(len(images))] // Update image
				}
			}

			// Write an update event only if there has been a change
			if vmType != originalType || state != originalState || image != originalImage {
				writeEvent(db, lastUpdateTime, projectID, vmID, vmType, state, image)
			}
		}

		if randGen.Float32() < 0.9 { // 90% chance of deletion
			deleteTime := randomDateAfter(lastUpdateTime, randGen)
			writeEvent(db, deleteTime, projectID, vmID, vmType, "deleted", image)
		}
	}
}

func writeEvent(db *gorm.DB, timestamp time.Time, projectID, vmID uuid.UUID, vmType, state string, image uuid.UUID) {
	instanceEvent := models.InstanceEvent{
		Timestamp: timestamp,
		ProjectID: types.BinaryUUID(projectID),
		UUID:      types.BinaryUUID(vmID),
		Type:      vmType,
		State:     models.InstanceState(state),
		Image:     types.BinaryUUID(image),
	}

	ret := db.Create(&instanceEvent)
	if ret.Error != nil {
		panic(ret.Error)
	}
}

func randomDateAfter(start time.Time, randGen *rand.Rand) time.Time {
	min := start.Unix()
	max := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix() // Assuming you want the end date to be 2022-01-01
	delta := max - min

	sec := randGen.Int63n(delta) + min
	return time.Unix(sec, 0)
}
