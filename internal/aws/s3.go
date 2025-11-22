package aws

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/rk1165/loadsimulator/internal/load"
	"github.com/rk1165/loadsimulator/internal/logger"
	"github.com/rk1165/loadsimulator/internal/types"
)

type LoadS3 struct {
	load.BaseLoad
	log       load.Log
	bucket    string
	key       string
	extension string
	body      string
	client    *s3.Client
	region    string
}

func NewS3(s3Config types.S3Config, cfg types.Config, client *s3.Client) *LoadS3 {
	s3Load := &LoadS3{
		bucket:    s3Config.Bucket,
		key:       s3Config.Key,
		extension: s3Config.Extension,
		region:    s3Config.Region,
		body:      s3Config.ResolveBody(),
		BaseLoad:  load.NewBaseLoad(cfg),
		log:       logger.CreateLoadLog(cfg.Name),
		client:    client,
	}
	s3Load.log.InfoLog.Printf("Initialized S3Load configs successfully")
	return s3Load
}

func (s *LoadS3) Execute(ctx context.Context, id uint64) error {
	start := time.Now()
	key := fmt.Sprintf("%s/%s%s", s.key, uuid.New().String(), s.extension)
	resp, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(s.body)),
	})
	if err != nil {
		return err
	}
	duration := time.Since(start)
	if s.Success(resp) {
		s.Record(duration, true)
		s.log.InfoLog.Printf("[S3 UPLOAD] requestId=%d etag=%s elapsed=%s", id, aws.ToString(resp.ETag), duration)
	} else {
		s.Record(duration, false)
		s.log.ErrorLog.Printf("[S3 UPLOAD] requestId=%d etag=%s elapsed=%s", id, aws.ToString(resp.ETag), duration)
	}
	return nil
}

func (s *LoadS3) Success(response any) bool {
	s3Response := response.(*s3.PutObjectOutput)
	return s3Response.ETag != nil
}

func (s *LoadS3) CalculateStats() *load.Stats {
	return s.BaseLoad.CalculateStats()
}
