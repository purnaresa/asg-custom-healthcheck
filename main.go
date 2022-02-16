package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	appName            string
	appUrl             string
	unhealthyCounter   int
	healthyCounter     int
	unhealthyThreshold int
	healtyhThreshold   int
)

func init() {
	appName = "dumdumapp"
	appUrl = "http://localhost:8080/healthcheck"
	unhealthyCounter = 0
	healthyCounter = 0
	unhealthyThreshold = 3
	healtyhThreshold = 3
	log.SetLevel(log.DebugLevel)
}

func main() {
	for {
		fileOk, err := os.OpenFile(fmt.Sprintf("%s_healthcheck.log", appName), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
		}
		defer fileOk.Close()

		log.SetOutput(fileOk)

		resp, err := http.Get(appUrl)
		if err != nil {
			unhealthyCounter++
			healthyCounter = 0

			log.WithFields(log.Fields{
				"status":            err.Error(),
				"unhealthy counter": unhealthyCounter,
			}).Error("healthcheck")

			if unhealthyCounter == unhealthyThreshold {
				setUnhealthy()
			}
		} else {
			if resp.StatusCode == http.StatusOK {
				healthyCounter++
				unhealthyCounter = 0

				log.WithFields(log.Fields{
					"healthy counter": healthyCounter,
				}).Info("healthcheck")

				if healthyCounter == healtyhThreshold {
					healthyCounter = 0
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func setUnhealthy() {
	log.Info("setting instance to unhealthy")
}
