package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
)

type LoadSQS struct {
	load.BaseLoad
	log      load.Log
	queueUrl string
	body     string
	attrs    map[string]sqsTypes.MessageAttributeValue
	client   *sqs.Client
	region   string
}

func NewSqs(sqsConfig types.SqsConfig, cfg types.Config, client *sqs.Client) *LoadSQS {
	sqsLoad := &LoadSQS{
		BaseLoad: load.NewBaseLoad(cfg),
		client:   client,
		log:      logger.CreateLoadLog(cfg.Name),
		body:     sqsConfig.ResolveBody(),
		region:   sqsConfig.Region,
		attrs:    buildMessageAttributes(sqsConfig.MessageAttributes),
	}

	out, err := client.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(sqsConfig.Queue),
	})
	if err != nil {
		return nil
	}
	sqsLoad.queueUrl = aws.ToString(out.QueueUrl)
	sqsLoad.log.InfoLog.Printf("Initialized SQSLoad configs successfully")
	return sqsLoad
}

func buildMessageAttributes(attrs []types.MessageAttribute) map[string]sqsTypes.MessageAttributeValue {
	messageAttributes := make(map[string]sqsTypes.MessageAttributeValue)

	for _, attr := range attrs {
		switch attr.Type {
		case "String":
			messageAttributes[attr.Name] = sqsTypes.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(attr.Value),
			}
		case "Number":
			messageAttributes[attr.Name] = sqsTypes.MessageAttributeValue{
				DataType:    aws.String("Number"),
				StringValue: aws.String(attr.Value),
			}
		case "Binary":
			messageAttributes[attr.Name] = sqsTypes.MessageAttributeValue{
				DataType:    aws.String("Binary"),
				BinaryValue: []byte(attr.Value),
			}
		default:
			messageAttributes[attr.Name] = sqsTypes.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(attr.Value),
			}
		}
	}
	return messageAttributes
}

func (s *LoadSQS) Execute(ctx context.Context, id uint64) error {
	start := time.Now()
	out, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(s.queueUrl),
		MessageBody:       aws.String(s.body),
		MessageAttributes: s.attrs,
	})
	if err != nil {
		return err
	}
	duration := time.Since(start)
	if s.Success(out) {
		s.Record(duration, true)
		s.log.InfoLog.Printf("[SQS SEND] requestId=%d messageId=%s elapsed=%s", id, aws.ToString(out.MessageId), duration)
	} else {
		s.Record(duration, false)
		s.log.ErrorLog.Printf("[SQS SEND] requestId=%d messageId=%s elapsed=%s", id, aws.ToString(out.MessageId), duration)
	}
	return nil
}

func (s *LoadSQS) Success(response any) bool {
	sqsResponse := response.(*sqs.SendMessageOutput)
	return sqsResponse.MessageId != nil
}

func (s *LoadSQS) CalculateStats() *load.Stats {
	return s.BaseLoad.CalculateStats()
}
