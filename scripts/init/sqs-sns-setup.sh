#!/bin/bash

echo "🚀 Bootstrapping SNS + SQS..."

TOPIC_ARN=$(awslocal sns create-topic \
  --name my-topic \
  --query 'TopicArn' \
  --output text)

QUEUE_URL=$(awslocal sqs create-queue \
  --queue-name my-queue \
  --query 'QueueUrl' \
  --output text)

QUEUE_ARN=$(awslocal sqs get-queue-attributes \
  --queue-url $QUEUE_URL \
  --attribute-names QueueArn \
  --query 'Attributes.QueueArn' \
  --output text)

# Policy restricting ONLY this topic (correct way)
awslocal sqs set-queue-attributes \
  --queue-url $QUEUE_URL \
  --attributes "Policy={
    \"Version\":\"2012-10-17\",
    \"Statement\":[{
      \"Effect\":\"Allow\",
      \"Principal\":{\"Service\":\"sns.amazonaws.com\"},
      \"Action\":\"sqs:SendMessage\",
      \"Resource\":\"$QUEUE_ARN\",
      \"Condition\":{\"ArnEquals\":{\"aws:SourceArn\":\"$TOPIC_ARN\"}}
    }]
  }"

awslocal sns subscribe \
  --topic-arn $TOPIC_ARN \
  --protocol sqs \
  --notification-endpoint $QUEUE_ARN \
  --attributes RawMessageDelivery=true

echo "Creating S3 bucket..."
awslocal s3 mb s3://localstack-bucket
echo "S3 bucket ready"

echo "🔐 Creating KMS key..."
KEY_ID=$(awslocal kms create-key \
  --query 'KeyMetadata.KeyId' \
  --output text)
echo "KMS Key ID: $KEY_ID"

# Optional alias
awslocal kms create-alias \
  --alias-name alias/localstack-kms \
  --target-key-id $KEY_ID

echo "✅ Infrastructure is ready"
