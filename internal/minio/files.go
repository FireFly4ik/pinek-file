package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
	"time"
)

type FileDataType struct {
	FileName  string
	FileOwner string
	Data      *bytes.Reader
}

func (m *minioClient) CreateOne(file FileDataType) (string, string, error) {
	objectID := uuid.New().String()
	contentType := "image/jpeg"
	if strings.Split(file.FileName, "."); len(strings.Split(file.FileName, ".")) > 1 {
		extension := strings.ToLower(strings.Split(file.FileName, ".")[1])
		switch extension {
		case "png":
			contentType = "image/png"
		case "gif":
			contentType = "image/gif"
		}
		objectID += "." + extension
	}

	m.wg.Add(1)
	defer m.wg.Done()

	// Загрузка данных в бакет Minio с использованием контекста для возможности отмены операции.
	_, err := m.mc.PutObject(context.Background(), m.bucketName, objectID, file.Data, file.Data.Size(), minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"File-Owner": file.FileOwner,
		},
	})
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

func (m *minioClient) GetOne(objectID string) (string, error) {
	// Получение предварительно подписанного URL для доступа к объекту Minio.
	url, err := m.mc.PresignedGetObject(context.Background(), m.bucketName, objectID, time.Second*24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении URL для объекта %s: %v", objectID, err)
	}

	return url.String(), nil
}

func (m *minioClient) GetMany(objectIDs []string) []string {
	var wg sync.WaitGroup
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, _ := range objectIDs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url, err := m.GetOne(objectIDs[i])
			if err != nil {
				log.Error().Err(err).Msgf("ошибка при получении объекта %s", objectIDs[i])
				return
			}
			objectIDs[i] = url
		}(i)
	}

	wg.Wait()

	return objectIDs
}

func (m *minioClient) DeleteOne(objectID string, fileOwner string) error {
	// Получение метаданных объекта для проверки владельца.
	objInfo, err := m.mc.StatObject(context.Background(), m.bucketName, objectID, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка при получении информации об объекте %s: %v", objectID, err)
	}

	// Проверка, совпадает ли владелец файла с предоставленным.
	if owner, ok := objInfo.UserMetadata["File-Owner"]; !ok || owner != fileOwner {
		return fmt.Errorf("пользователь не является владельцем файла")
	}

	// Удаление объекта из бакета Minio.
	err = m.mc.RemoveObject(context.Background(), m.bucketName, objectID, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *minioClient) DeleteMany(objectIDs []string, fileOwner string) bool {
	var wg sync.WaitGroup

	errors := false

	for _, objectID := range objectIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			err := m.DeleteOne(id, fileOwner)
			if err != nil {
				log.Error().Err(err).Msgf("ошибка при удалении объекта %s", id)
				errors = true
			}
		}(objectID)
	}

	wg.Wait()
	return errors
}
