package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	region             string
	sqsEndpoint        string
	appName            string
	appUrl             string
	unhealthyCounter   int
	healthyCounter     int
	unhealthyThreshold int
	healthyThreshold   int
	interval           int
	logLevel           string
	logFileWrite       bool
)

// fetch config from config.yaml
func getConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	errConfig := viper.ReadInConfig()
	if errConfig != nil {
		log.Fatalln(errConfig)
	} else if viper.GetString("APPNAME") == "" {
		log.Fatalln("NO APP NAME")
	}

	region = viper.GetString("REGION")
	sqsEndpoint = viper.GetString("SQS_ENDPOINT")
	appName = viper.GetString("APPNAME")
	appUrl = viper.GetString("APPURL")
	unhealthyThreshold = viper.GetInt("HEALTHY_THRESHOLD")
	healthyThreshold = viper.GetInt("UNHEALTHY_THRESHOLD")
	interval = viper.GetInt("INTERVAL")
	logLevel = viper.GetString("LOG_LEVEL")
	logFileWrite = viper.GetBool("LOG_FILE_WRITE")

	log.WithField("appName", appName).Info("get config complete")
}

func init() {
	unhealthyCounter = 0
	healthyCounter = 0

	// apply log config
	getConfig()
	level, errLevel := log.ParseLevel(logLevel)
	if errLevel != nil {
		log.Fatalln(errLevel)
	}
	log.SetLevel(level)
	//
}

func main() {
	// set log file
	if logFileWrite == true {
		fileOk, err := os.OpenFile(fmt.Sprintf("%s_healthcheck.log", appName), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
		}
		defer fileOk.Close()

		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(fileOk)
	}
	//

	for {
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

				log.WithField("healthy counter", healthyCounter).
					Debug("healthcheck")

				if healthyCounter == healthyThreshold {
					setHealthy()
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func setHealthy() {
	healthyCounter = 0
}

func setUnhealthy() (messageID string, err error) {
	// AWS Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
		}),
		config.WithDefaultRegion(region))
	if err != nil {
		log.Errorln(err)
		return
	}
	//

	// get instance id
	clientEC2 := imds.NewFromConfig(cfg)
	inputImds := imds.GetInstanceIdentityDocumentInput{}
	outputImds, err := clientEC2.GetInstanceIdentityDocument(context.Background(),
		&inputImds)
	if err != nil {
		log.Errorln(err)
		return
	}
	instanceID := outputImds.InstanceIdentityDocument.InstanceID
	//

	// send SQS message
	sqsClient := sqs.NewFromConfig(cfg)
	queueUrl := aws.String(sqsEndpoint)
	inputSend := sqs.SendMessageInput{
		MessageBody: &instanceID,
		QueueUrl:    queueUrl}
	ouputSend, err := sqsClient.SendMessage(context.Background(),
		&inputSend)
	if err != nil {
		log.Errorln(err)
		return
	}
	messageID = *ouputSend.MessageId
	//

	log.WithField("message_id", messageID).Info("set unhealty message sent")
	return

}
