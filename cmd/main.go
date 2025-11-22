package main

import (
	"context"
	"flag"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rk1165/loadsimulator/internal/aws"
	"github.com/rk1165/loadsimulator/internal/config"
	"github.com/rk1165/loadsimulator/internal/kafka"
	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/rest"
	"github.com/rk1165/loadsimulator/internal/types"
)

func main() {

	configType := flag.String("configType", "", "The type of config to load")
	subConfig := flag.String("subConfig", "", "The type of subconfig to load")
	scenarioName := flag.String("scenario", "", "The name of the scenario to load")
	flag.Parse()

	logger.InfoLog.Printf("configType=%s scenarioName=%s\n", *configType, *scenarioName)

	configFile := fmt.Sprintf("configs/%s.yaml", *subConfig)
	ctx := context.Background()

	ch := make(chan *load.Stats, 1)
	var runner *load.Runner
	switch *configType {
	case "rest":
		apiConfig, cfg, err := config.LoadTestConfig[types.ApiConfig](configFile, *scenarioName)
		if err != nil {
			logger.ErrorLog.Fatal(err)
		}
		var httpLoad load.Load
		if *subConfig == "post" {
			httpLoad = rest.NewPost(apiConfig, cfg)
		} else {
			httpLoad = rest.NewGet(apiConfig, cfg)
		}
		runner = load.NewLoadRunner(httpLoad, cfg)
	case "aws":
		if *subConfig == "s3" {
			s3Config, cfg, err := config.LoadTestConfig[types.S3Config](configFile, *scenarioName)
			if err != nil {
				logger.ErrorLog.Fatal(err)
			}
			awsS3Config, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(s3Config.Region))
			if err != nil {
				logger.ErrorLog.Fatal(err)
			}
			s3Client := s3.NewFromConfig(awsS3Config)
			s3Load := aws.NewS3(s3Config, cfg, s3Client)
			runner = load.NewLoadRunner(s3Load, cfg)
		} else if *subConfig == "sqs" {
			sqsConfig, cfg, err := config.LoadTestConfig[types.SqsConfig](configFile, *scenarioName)
			if err != nil {
				logger.ErrorLog.Fatal(err)
			}
			awsSqsConfig, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(sqsConfig.Region))
			if err != nil {
				logger.ErrorLog.Fatal(err)
			}
			sqsClient := sqs.NewFromConfig(awsSqsConfig)
			sqsLoad := aws.NewSqs(sqsConfig, cfg, sqsClient)
			if sqsLoad == nil {
				logger.ErrorLog.Fatal("unable to initialize sqs load")
			}
			runner = load.NewLoadRunner(sqsLoad, cfg)
		}
	case "kafka":
		kafkaConfig, cfg, err := config.LoadTestConfig[types.KafkaConfig](configFile, *scenarioName)
		if err != nil {
			logger.ErrorLog.Fatal(err)
		}
		kafkaLoad := kafka.NewKafka(kafkaConfig, cfg)
		if kafkaLoad == nil {
			logger.ErrorLog.Fatal("unable to initialize kafka load")
		}
		logger.InfoLog.Printf("Kafka Loader Initialized")
		runner = load.NewLoadRunner(kafkaLoad, cfg)
	}
	if runner != nil {
		if err := runner.Run(ctx, ch); err != nil {
			logger.ErrorLog.Fatalf("load run failed for load=%s/%s scenario=%s with error=[%v]",
				*configType, *subConfig, *scenarioName, err)
		}
		stats := <-ch
		logger.InfoLog.Printf("Load finished successfully for scenario: %s, stats: %+v", *scenarioName, stats)
	} else {
		logger.ErrorLog.Fatal(fmt.Errorf("unknown configType: %s", *configType))
	}
}
