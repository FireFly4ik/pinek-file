package minio

import (
	"context"
	"file/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

// Client интерфейс для взаимодействия с Minio
type Client interface {
	CreateOne(file FileDataType) (string, string, error)
	CreateMany([]FileDataType) ([]string, []string, error)
	GetOne(objectID string) (string, error)
	GetMany(objectIDs []string) []string
	DeleteOne(objectID string) error
	DeleteMany(objectIDs []string)
	Close()
}

type minioClient struct {
	mc         *minio.Client
	bucketName string
	wg         sync.WaitGroup
}

func NewMinioClient(envConf *config.Config) Client {
	ctx := context.Background()

	minioAddress := envConf.Minio.IP + ":" + envConf.Minio.Port

	client, err := minio.New(minioAddress, &minio.Options{
		Creds:  credentials.NewStaticV4(envConf.Minio.RootUser, envConf.Minio.RootPassword, ""),
		Secure: envConf.Minio.UseSSL == "true",
	})
	if err != nil {
		panic("failed to connect to minio: " + err.Error())
	}

	m := &minioClient{
		mc:         client,
		bucketName: envConf.Minio.BucketName,
		wg:         sync.WaitGroup{},
	}

	exists, err := m.mc.BucketExists(ctx, envConf.Minio.BucketName)
	if err != nil {
		panic("failed to check if bucket exists: " + err.Error())
	}
	if !exists {
		err := m.mc.MakeBucket(ctx, envConf.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			panic("failed to create bucket: " + err.Error())
		}
	}

	return m
}

func (m *minioClient) Close() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	done := make(chan struct{})

	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("minio shutdown complete")
	case <-ctx.Done():
		log.Warn().Msg("minio shutdown timeout reached")
	}
}
