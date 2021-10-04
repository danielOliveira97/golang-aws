package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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

func createQueueIfNotExist(sqsClient *sqs.SQS) *sqs.CreateQueueOutput {
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "HelloWorldTestQueue"
	}

	req, q := sqsClient.CreateQueueRequest(&sqs.CreateQueueInput{
		QueueName: &queueName,
	})

	req.SetContext(context.Background())
	if err := req.Send(); req != nil {
		fmt.Println("create queue:", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		os.Exit(1)
	}()

	return q
}

func SendSQSMessage(sqsClient *sqs.SQS, messageBody string) {
	queueUrl := createQueueIfNotExist(sqsClient).QueueUrl

	for i := 1; i < 20000; i++ {
		fmt.Println("sending message ", i)
		message := fmt.Sprintf("%s %d", messageBody, i)
		req, resp := sqsClient.SendMessageRequest(&sqs.SendMessageInput{
			MessageBody: &message,
			QueueUrl:    queueUrl,
		})
		req.SetContext(context.Background())
		if err := req.Send(); err != nil {
			fmt.Println("Error", err)
			return
		}

		fmt.Println("Success", aws.StringValue(resp.MessageId))
		//		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}

func main() {
	SendSQSMessage(loadAwsConfig(), *aws.String("Mensagem de teste"))
}
