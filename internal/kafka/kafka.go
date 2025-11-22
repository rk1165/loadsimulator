package kafka

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	"github.com/rk1165/loadsimulator/internal"
	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/oauth"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

type LoadKafka struct {
	load.BaseLoad
	log    load.Log
	topic  string
	client *kgo.Client
	body   string
}

func NewKafka(kafkaConfig types.KafkaConfig, cfg types.Config) *LoadKafka {
	var mechanism sasl.Mechanism
	if kafkaConfig.Authentication == "oauth" {
		mechanism = getOauthMechanism(kafkaConfig)
	} else if kafkaConfig.Authentication == "scram" {
		mechanism = getScramMechanism(kafkaConfig)
	} else {
		logger.ErrorLog.Fatalf("Invalid authentication type %s", kafkaConfig.Authentication)
	}
	// let TLS not verify Kafka broker's certificate
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaConfig.Broker),
		kgo.SASL(mechanism),
		kgo.DialTLSConfig(tlsConfig),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.DisableIdempotentWrite(),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		logger.ErrorLog.Fatalf("Unable to initialize kafka client %v", err)
	}
	// ping the broker to see if it's reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = client.Ping(ctx); err != nil {
		logger.ErrorLog.Printf("Kafka broker unreachable: %v", err)
		return nil
	}

	kafkaLoad := &LoadKafka{
		BaseLoad: load.NewBaseLoad(cfg),
		client:   client,
		log:      logger.CreateLoadLog(cfg.Name),
		body:     kafkaConfig.ResolveBody(),
		topic:    kafkaConfig.Topic,
	}
	kafkaLoad.log.InfoLog.Printf("Initialized KafkaLoad configs successfully")
	return kafkaLoad
}

func getOauthMechanism(kafkaConfig types.KafkaConfig) sasl.Mechanism {
	oauthConfig := &internal.OAuthBearer{
		ClientId:     kafkaConfig.ClientId,
		ClientSecret: kafkaConfig.ClientSecret,
		TokenUrl:     internal.OauthUrl,
	}
	token, err := oauthConfig.GenerateToken()
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	return oauth.Auth{
		Token: token,
	}.AsMechanism()
}

func getScramMechanism(kafkaConfig types.KafkaConfig) sasl.Mechanism {
	scramAuth := scram.Auth{
		User: kafkaConfig.UserName,
		Pass: kafkaConfig.Password,
	}
	return scramAuth.AsSha512Mechanism()
}

func (k *LoadKafka) Execute(ctx context.Context, id uint64) error {
	start := time.Now()
	rec := &kgo.Record{
		Topic: k.topic,
		Value: []byte(k.body),
	}
	results := k.client.ProduceSync(ctx, rec)
	duration := time.Since(start)
	if k.Success(results) {
		k.Record(duration, true)
		k.log.InfoLog.Printf("[Kafka Producer] requestId=%d partition=%d offset=%d", id, results[0].Record.Partition, results[0].Record.Offset)
	} else {
		k.Record(duration, false)
		k.log.ErrorLog.Printf("[Kafka Producer] requestId=%d", id)
	}
	return nil
}

func (k *LoadKafka) Success(response any) bool {
	results := response.(kgo.ProduceResults)
	if err := results.FirstErr(); err != nil {
		return false
	}
	return true
}

func (k *LoadKafka) CalculateStats() *load.Stats {
	defer k.client.Close()
	return k.BaseLoad.CalculateStats()
}
