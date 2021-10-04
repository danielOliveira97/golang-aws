package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func loadAwsConfig() *sqs.SQS {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	cfg := aws.Config{
		Region: aws.String(region),
	}
	sess := session.Must(session.NewSession(&cfg))
	svc := sqs.New(sess)
	return svc
}

func getQueueUrl(sqsClient *sqs.SQS) string {
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "HelloWorldTestQueue"
	}

	fmt.Println("listening to queue:", queueName)
	req, q := sqsClient.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	req.SetContext(context.Background())
	if err := req.Send(); req != nil {
		// handle queue creation error
		fmt.Println("cannot get queue:", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		os.Exit(1)
	}()

	queueUrl := q.QueueUrl
	return *queueUrl
}

func ConsumerSQSMessage(sqsClient *sqs.SQS) {
	timeout := int64(20)
	queueUrl := aws.String(getQueueUrl(sqsClient))

	for {
		req, msg := sqsClient.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
			QueueUrl:        queueUrl,
			WaitTimeSeconds: &timeout,
		})
		req.SetContext(context.Background())
		if err := req.Send(); err == nil {
			fmt.Println("message:", msg)
		} else {
			fmt.Println("error receiving message from queue:", err)
		}
		if len(msg.Messages) > 0 {
			req, _ := sqsClient.DeleteMessageRequest(&sqs.DeleteMessageInput{
				QueueUrl:      queueUrl,
				ReceiptHandle: msg.Messages[0].ReceiptHandle,
			})
			req.SetContext(context.Background())
			if err := req.Send(); err != nil {
				fmt.Println("error deleting message from queue:", err)
			}
		}
		// Implement some delay here to simulate processing time
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}
}

func main() {
	ConsumerSQSMessage(loadAwsConfig())
}
