package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type FileDataType struct {
	FileName string
	Data     []byte
}

func (m *minioClient) CreateOne(file FileDataType) (string, string, error) {
	objectID := uuid.New().String()

	reader := bytes.NewReader(file.Data)

	m.wg.Add(1)
	defer m.wg.Done()

	// Загрузка данных в бакет Minio с использованием контекста для возможности отмены операции.
	_, err := m.mc.PutObject(context.Background(), m.bucketName, objectID, reader, int64(len(file.Data)), minio.PutObjectOptions{})
	if err != nil {
		return "", "", fmt.Errorf("ошибка при создании объекта %s: %v", file.FileName, err)
	}

	// Получение URL для загруженного объекта
	url, err := m.mc.PresignedGetObject(context.Background(), m.bucketName, objectID, time.Second*24*60*60, nil)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при создании URL для объекта %s: %v", file.FileName, err)
	}

	return objectID, url.String(), nil
}

func (m *minioClient) CreateMany(data []FileDataType) ([]string, []string, error) {
	urls := make([]string, 0, len(data))
	uuids := make([]string, 0, len(data))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filesCh := make(chan struct {
		ObjectID string
		URL      string
	}, len(data))

	var wg sync.WaitGroup

	for _, file := range data {
		m.wg.Add(1)
		wg.Add(1)

		go func(file FileDataType) {
			defer wg.Done()
			defer m.wg.Done()

			objectID := uuid.New().String()

			// Загрузка данных в бакет Minio
			_, err := m.mc.PutObject(ctx, m.bucketName, objectID, bytes.NewReader(file.Data), int64(len(file.Data)), minio.PutObjectOptions{})
			if err != nil {
				cancel()
				return
			}

			// Получение URL для загруженного объекта
			url, err := m.mc.PresignedGetObject(ctx, m.bucketName, objectID, time.Second*24*60*60, nil)
			if err != nil {
				cancel()
				return
			}

			combo := struct {
				ObjectID string
				URL      string
			}{
				ObjectID: objectID,
				URL:      url.String(),
			}

			filesCh <- combo
		}(file)
	}

	if ctx.Err() != nil {
		return nil, nil, fmt.Errorf("ошибка при создании объектов: операция была отменена")
	}

	wg.Wait()
	close(filesCh)

	for combo := range filesCh {
		urls = append(urls, combo.URL)
		uuids = append(uuids, combo.ObjectID)
	}

	return urls, uuids, nil
}

func (m *minioClient) GetOne(objectID string) (string, error) {
	// Получение предварительно подписанного URL для доступа к объекту Minio.
	url, err := m.mc.PresignedGetObject(context.Background(), m.bucketName, objectID, time.Second*24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL для объекта %s: %v", objectID, err)
	}

	return url.String(), nil
}

func (m *minioClient) GetMany(objectIDs []string) []string {
	urls := make([]string, 0, len(objectIDs))

	var wg sync.WaitGroup
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, objectID := range objectIDs {
		wg.Add(1)
		go func(i int, objectID string) {
			defer wg.Done()
			url, err := m.GetOne(objectID)
			if err != nil {
				log.Error().Err(err).Msgf("ошибка при получении объекта %s", objectID)
				return
			}
			urls[i] = url
		}(i, objectID)
	}

	wg.Wait()

	return urls
}

func (m *minioClient) DeleteOne(objectID string) error {
	// Удаление объекта из бакета Minio.
	err := m.mc.RemoveObject(context.Background(), m.bucketName, objectID, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *minioClient) DeleteMany(objectIDs []string) {
	var wg sync.WaitGroup

	for _, objectID := range objectIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			err := m.mc.RemoveObject(context.Background(), m.bucketName, id, minio.RemoveObjectOptions{})
			if err != nil {
				log.Error().Err(err).Msgf("ошибка при удалении объекта %s", id)
			}
		}(objectID)
	}

	wg.Wait()
}
