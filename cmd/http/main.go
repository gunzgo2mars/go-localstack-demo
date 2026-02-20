package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gunzgo2mars/go-localstack-demo/pkg/awsconfig"
)

func main() {

	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	topicArn := os.Getenv("TOPIC_ARN")

	awsConfig, err := awsconfig.InitAwsConfig(endpoint, region)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	client := sns.NewFromConfig(awsConfig)

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {

		message := r.URL.Query().Get("message")

		client.Publish(context.TODO(), &sns.PublishInput{
			TopicArn: aws.String(topicArn),
			Message:  aws.String(message),
		})

		w.Write([]byte("message sent"))

	})

	http.ListenAndServe(":7777", nil)
}
