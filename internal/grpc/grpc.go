package grpc

import (
	"file/internal/config"
	"file/internal/minio"
	pb "file/internal/proto"
)

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	envConf *config.Config
	minio   minio.Client
}

func NewFileServer(cfg *config.Config, minioClient minio.Client) *FileServiceServer {
	return &FileServiceServer{
		envConf: cfg,
		minio:   minioClient,
	}
}
