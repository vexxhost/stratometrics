package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
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
	states    = []string{"active", "stopped", "suspened", "error"}
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

	file, err := os.Create("testdata/instance_events.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i := 0; i < numVMs; i++ {
		projectID := projects[randGen.Intn(numProjects)] // Randomly pick a project UUID
		vmID := uuid.New()
		vmType := fmt.Sprintf("type-%d", randGen.Intn(numTypes)+1)
		initialStates := []string{"active", "building"}
		state := initialStates[randGen.Intn(len(initialStates))] // Randomly pick initial state
		image := images[randGen.Intn(numImages)]                 // Randomly pick an image UUID
		createTime := randomDateAfter(startDate, randGen)

		writeEvent(writer, createTime, projectID, vmID, vmType, state, image)

		// If initial state is "BUILD", ensure a follow-up "UPDATE" to "ACTIVE" or "ERROR"
		lastUpdateTime := createTime
		if state == "BUILD" {
			lastUpdateTime = randomDateAfter(lastUpdateTime, randGen)
			updatedState := "ACTIVE"     // Default to "ACTIVE"
			if randGen.Float32() < 0.1 { // 10% chance to go to "ERROR"
				updatedState = "ERROR"
			}
			writeEvent(writer, lastUpdateTime, projectID, vmID, vmType, updatedState, image)
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
				writeEvent(writer, lastUpdateTime, projectID, vmID, vmType, state, image)
			}
		}

		if randGen.Float32() < 0.9 { // 90% chance of deletion
			deleteTime := randomDateAfter(lastUpdateTime, randGen)
			writeEvent(writer, deleteTime, projectID, vmID, vmType, "deleted", image)
		}
	}
}

func writeEvent(writer *bufio.Writer, timestamp time.Time, projectID, vmID uuid.UUID, vmType, state string, image uuid.UUID) {
	_, err := writer.WriteString(
		fmt.Sprintf(
			"%s,%s,%s,%s,%s,%s\n",
			timestamp.Format(dateFormat), projectID, vmID, vmType, state, image,
		),
	)
	if err != nil {
		panic(err)
	}
}

func randomDateAfter(start time.Time, randGen *rand.Rand) time.Time {
	min := start.Unix()
	max := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix() // Assuming you want the end date to be 2022-01-01
	delta := max - min

	sec := randGen.Int63n(delta) + min
	return time.Unix(sec, 0)
}
