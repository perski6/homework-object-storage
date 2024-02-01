package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/perski6/homework-object-storage/config"
	"github.com/perski6/homework-object-storage/consistentHash"
)

type Service struct {
	nodeProvider consistentHash.NodeProvider[*minio.Client]
	logger       slog.Logger
}

var InstanceNotAccessible = errors.New("cannot access minio instance")

func New(provider consistentHash.NodeProvider[*minio.Client]) *Service {
	return &Service{
		nodeProvider: provider,
		logger:       slog.Logger{},
	}
}

func (c *Service) GetObject(ctx context.Context, id string) (body []byte, err error) {
	node := c.nodeProvider.PickNode(id)
	if node.Status == consistentHash.Stopped {
		return nil, InstanceNotAccessible
	}
	client := node.Client

	object, err := client.GetObject(ctx, config.App.Bucket, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	// Read the object content.
	return readObjectContent(object)
}

func (c *Service) PutObject(ctx context.Context, id string, body []byte) error {
	node := c.nodeProvider.PickNode(id)
	if node.Status == consistentHash.Stopped {
		return InstanceNotAccessible
	}
	client := node.Client

	_, err := client.PutObject(ctx, config.App.Bucket, id, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func readObjectContent(object *minio.Object) ([]byte, error) {
	var response bytes.Buffer
	fileBody := make([]byte, 1024)

	for {
		n, err := object.Read(fileBody)
		if err != nil && err != io.EOF {
			return nil, err
		}
		response.Write(fileBody[:n])
		if err == io.EOF {
			break
		}
	}

	return response.Bytes(), nil
}
