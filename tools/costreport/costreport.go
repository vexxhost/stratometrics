package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Rhymond/go-money"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/vexxhost/stratometrics/internal/api/v1alpha1/handlers"
)

type Config struct {
	InstanceTypes map[string]InstanceType `yaml:"instance_types"`
}

type InstanceType struct {
	Cost       float64                `yaml:"cost"`
	Equivilant InstanceTypeEquivilant `yaml:"equivilant"`
}

type InstanceTypeEquivilant struct {
	AWS InstanceTypeEquivilantInfo `yaml:"aws"`
}

type InstanceTypeEquivilantInfo struct {
	Name string  `yaml:"name"`
	Cost float64 `yaml:"cost"`
}

func main() {
	configFile, err := os.Open("./tools/costreport/config.yml")
	if err != nil {
		log.WithError(err).Fatal("could not open config file")
	}
	defer configFile.Close()

	var config Config
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		log.WithError(err).Fatal("could not unmarshal config file")
	}

	if err := godotenv.Load(); err != nil {
		log.WithError(err).Warn("could not load .env file")
	}

	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.WithError(err).Fatal("could not load auth options from env")
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		log.WithError(err).Fatal("could not create authenticated client")
	}

	token := provider.Token()

	projectId := "32f6b15efafe477b9f3f378926069547"

	url := fmt.Sprintf("%s/v1alpha1/instances", os.Getenv("STRATOMETRICS_ENDPOINT"))
	if projectId != "" {
		url = fmt.Sprintf("%s?project_id=%s", url, projectId)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Fatal("could not create request")
	}

	req.Header.Set("X-Auth-Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Fatal("could not make request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Fatal("could not read response body")
	}

	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"status": resp.StatusCode,
			"body":   body,
		}).Fatal("unexpected response")
	}

	var instancesResponse handlers.InstancesResponse
	if err := yaml.Unmarshal(body, &instancesResponse); err != nil {
		log.WithError(err).Fatal("could not unmarshal response")
	}

	tableData := pterm.TableData{
		{"Instance Type", "Hours", "Cost", "AWS Equivilant", "AWS Cost"},
	}

	totalCost := money.New(0, money.USD)
	totalAwsCost := money.New(0, money.USD)

	for _, usage := range instancesResponse.Results {
		duration := usage.GetDuration()
		instanceType := config.InstanceTypes[usage.Type]
		cost := money.NewFromFloat(instanceType.Cost*duration.Hours(), money.USD)
		awsCost := money.NewFromFloat(instanceType.Equivilant.AWS.Cost*duration.Hours(), money.USD)

		tableData = append(tableData, []string{
			usage.Type, duration.String(), cost.Display(), instanceType.Equivilant.AWS.Name, awsCost.Display(),
		})

		totalCost, err = totalCost.Add(cost)
		if err != nil {
			log.WithError(err).Fatal("could not add cost")
		}

		totalAwsCost, err = totalAwsCost.Add(awsCost)
		if err != nil {
			log.WithError(err).Fatal("could not add cost")
		}
	}

	pterm.DefaultHeader.
		WithMargin(15).
		WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("Instances")
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithBoxed().WithData(tableData).Render()

	fmt.Println("Total compute cost:", totalCost.Display())
	fmt.Println("Total AWS equivilant cost:", totalAwsCost.Display())

	savings, err := totalAwsCost.Subtract(totalCost)
	if err != nil {
		log.WithError(err).Fatal("could not subtract costs")
	}

	fmt.Println("Total savings:", savings.Display())
}
