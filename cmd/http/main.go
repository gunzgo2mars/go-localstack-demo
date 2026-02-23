package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gunzgo2mars/go-localstack-demo/pkg/awsconfig"
)

type EncryptRequest struct {
	Text string `json:"text"`
}

type EncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
}

type DecryptRequest struct {
	Ciphertext string `json:"ciphertext"`
}

type DecryptResponse struct {
	Text string `json:"text"`
}

func main() {

	ctx := context.Background()
	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	topicArn := os.Getenv("TOPIC_ARN")

	awsConfig, err := awsconfig.InitAwsConfig(endpoint, region)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	// SNS
	snsClient := sns.NewFromConfig(awsConfig)

	// S3
	s3Client := s3.NewFromConfig(
		awsConfig,
		func(o *s3.Options) {
			o.UsePathStyle = true
		},
	)
	uploader := transfermanager.New(
		s3Client,
		func(o *transfermanager.Options) {
			o.PartSizeBytes = 10 * 1024 * 1024
			o.Concurrency = 5
		},
	)

	// KMS
	kmsClient := kms.NewFromConfig(awsConfig)

	// SNS Publishing message handler
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {

		message := r.URL.Query().Get("message")

		snsClient.Publish(context.TODO(), &sns.PublishInput{
			TopicArn: aws.String(topicArn),
			Message:  aws.String(message),
		})

		w.Write([]byte("message sent"))
	})

	// S3 Uploader handler
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer file.Close()

		result, err := uploader.UploadObject(
			ctx,
			&transfermanager.UploadObjectInput{
				Bucket: aws.String("localstack-bucket"),
				Key:    aws.String(header.Filename),
				Body:   file,
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info(fmt.Sprintf("Uploaded: %s", *result.Key))
		w.Write([]byte("Upload successfully"))
	})

	// KMS Encryption handler
	http.HandleFunc("/encrypt", func(w http.ResponseWriter, r *http.Request) {
		var req EncryptRequest
		json.NewDecoder(r.Body).Decode(&req)

		out, err := kmsClient.Encrypt(ctx, &kms.EncryptInput{
			KeyId:     aws.String("alias/localstack-kms"),
			Plaintext: []byte(req.Text),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode binary ciphertext for JSON transport
		encoded := base64.StdEncoding.EncodeToString(out.CiphertextBlob)

		response := EncryptResponse{
			Ciphertext: encoded,
		}
		json.NewEncoder(w).Encode(response)
	})

	// KMS Decryption handler
	http.HandleFunc("/decrypt", func(w http.ResponseWriter, r *http.Request) {
		var req DecryptRequest
		json.NewDecoder(r.Body).Decode(&req)

		blob, err := base64.StdEncoding.DecodeString(req.Ciphertext)
		if err != nil {
			http.Error(w, "invalid base64", http.StatusBadRequest)
			return
		}

		out, err := kmsClient.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: blob,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := DecryptResponse{
			Text: string(out.Plaintext),
		}
		json.NewEncoder(w).Encode(response)
	})

	slog.Info("HTTP server running on :7777")
	http.ListenAndServe(":7777", nil)
}
