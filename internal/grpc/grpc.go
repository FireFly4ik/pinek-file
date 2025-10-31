package grpc

import (
	"file/internal/config"
	pb "file/internal/proto"
)

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	envConf *config.Config
}

func NewFileServer(cfg *config.Config) *FileServiceServer {
	return &FileServiceServer{
		envConf: cfg,
	}
}
