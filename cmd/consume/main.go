package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gunzgo2mars/go-localstack-demo/pkg/awsconfig"
)

func main() {

	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	queueURL := os.Getenv("QUEUE_URL")

	awsConfig, err := awsconfig.InitAwsConfig(endpoint, region)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	client := sqs.NewFromConfig(awsConfig)

	for {
		out, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 5,
			WaitTimeSeconds:     20,
		})

		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		for i, v := range out.Messages {
			fmt.Printf("MSG:INDEX:%d -> %s \n", i, *v.Body)

			client.DeleteMessage(
				context.TODO(),
				&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: v.ReceiptHandle,
				},
			)
		}
	}

}
