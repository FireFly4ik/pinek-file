package grpc

import (
	"bytes"
	"file/internal/config"
	"file/internal/minio"
	pb "file/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
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

func (fs *FileServiceServer) UploadFile(stream grpc.ClientStreamingServer[pb.UploadFileRequest, pb.UploadFileResponse]) error {
	var (
		fileName string
		buffer   bytes.Buffer
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break // клиент закончил передачу
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		if fileName == "" {
			fileName = req.FileName
		}

		_, err = buffer.Write(req.Data)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to write buffer: %v", err)
		}
	}

	file := minio.FileDataType{
		FileName: fileName,
		Data:     bytes.NewReader(buffer.Bytes()),
	}

	objectID, url, err := fs.minio.CreateOne(file)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to upload file: %v", err)
	}

	resp := &pb.UploadFileResponse{
		FileId:  objectID,
		FileUrl: url,
	}

	return stream.SendAndClose(resp)
}

//func (fs *FileServiceServer) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
//
//}
//
//func (fs *FileServiceServer) GetFiles(ctx context.Context, req *pb.GetFilesRequest) (*pb.GetFilesResponse, error) {
//
//}
//
//func (fs *FileServiceServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
//
//}
//
//func (fs *FileServiceServer) DeleteFiles(ctx context.Context, req *pb.DeleteFilesRequest) (*pb.DeleteFileResponse, error) {
//
//}
