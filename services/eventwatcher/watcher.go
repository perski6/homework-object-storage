package eventwatcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/perski6/homework-object-storage/config"
	"github.com/perski6/homework-object-storage/consistentHash"
)

const MINIOPREFIX string = "amazin-object-storage"
const MINIOUSER string = "MINIO_ROOT_USER"
const MINIOPASSWORD string = "MINIO_ROOT_PASSWORD"

type Instance struct {
	User     string
	Password string
	Host     string
}

func New(logger *slog.Logger, provider consistentHash.NodeProvider[*minio.Client]) *Watcher {
	w := &Watcher{
		logger:       logger,
		client:       client.Client{},
		nodeProvider: provider,
	}
	w.init()

	return w
}

type Watcher struct {
	logger       *slog.Logger
	client       client.Client
	nodeProvider consistentHash.NodeProvider[*minio.Client]
}

func (w *Watcher) DiscoverInstances() {
	var instances []Instance

	filterArgs := filters.NewArgs()
	filterArgs.Add("name", MINIOPREFIX+"*")
	containers, err := w.client.ContainerList(context.Background(), container.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		w.logger.Error("Error discovering: ", err)
	}

	for _, c := range containers {
		containerDetails, err := w.client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			w.logger.Error("error inspecting container: ", err)
			break
		}

		instance, err := extractStorageInstanceInfo(containerDetails)
		if err != nil {
			w.logger.Error("error getting container information: ", err)
			break
		}
		instances = append(instances, instance)
	}
	for _, instance := range instances {
		w.handleStartEvent(instance)
	}
}

func (w *Watcher) init() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	w.client = *dockerClient

	return nil
}

func (w *Watcher) Watch() {
	defer func() {
		if err := w.client.Close(); err != nil {
			w.logger.Error("error closing docker client", err)
		}
	}()

	options := types.EventsOptions{
		Filters: filters.NewArgs(),
	}

	eventsCh, errCh := w.client.Events(context.Background(), options)

	for {
		select {
		case event := <-eventsCh:
			if event.Type != events.ContainerEventType {
				continue
			}
			switch event.Action {
			case "start":
				{
					containerDetails, err := w.client.ContainerInspect(context.Background(), event.Actor.ID)
					if err != nil {
						w.logger.Error("error inspecting container: ", err)
					}
					instance, err := extractStorageInstanceInfo(containerDetails)
					if err != nil {
						w.logger.Error("error extracting instance info", err)
					}
					w.handleStartEvent(instance)
				}
			case "stop":
				{
					containerDetails, err := w.client.ContainerInspect(context.Background(), event.Actor.ID)
					if err != nil {
						w.logger.Error("error inspecting container: ", err)
					}
					instance, err := extractStorageInstanceInfo(containerDetails)
					if err != nil {
						w.logger.Error("error extracting instance info", err)
					}
					w.handleStopEvent(instance)
				}

			}

		case err := <-errCh:
			if err != nil {
				w.logger.Error("eventwatcher error: ", err)
			}

		}
	}

}

func (w *Watcher) handleStartEvent(instance Instance) {
	client, err := minio.New(instance.Host, &minio.Options{
		Creds: credentials.NewStaticV4(instance.User, instance.Password, ""),
	})
	err = client.MakeBucket(context.Background(), config.App.Bucket, minio.MakeBucketOptions{
		Region: "",
	})
	if err != nil {
		exists, bucketExistsErr := client.BucketExists(context.Background(), config.App.Bucket)
		if bucketExistsErr == nil && exists {
			w.logger.Info("bucket already exists")
		} else {
			w.logger.Error("bucket exists", err)
		}
	} else {
		w.logger.Info("created bucket", config.App.Bucket)
	}
	w.nodeProvider.AddNode(consistentHash.Node[*minio.Client]{
		Name:   instance.Host,
		Client: client,
	})
}

func (w *Watcher) handleStopEvent(instance Instance) {
	w.nodeProvider.RemoveNode(instance.Host)
}

func extractStorageInstanceInfo(inspect types.ContainerJSON) (Instance, error) {
	if !inspect.State.Running {
		return Instance{}, errors.New("container is not running")
	}

	user, err := getUserFromEnv(inspect.Config.Env)
	if err != nil {
		return Instance{}, err
	}

	pass, err := getPasswordFromEnv(inspect.Config.Env)
	if err != nil {
		return Instance{}, err
	}

	host, err := getContainerIPAddress(inspect)
	if err != nil {
		return Instance{}, err
	}

	hostWithPort := fmt.Sprintf("%s:%s", host, config.App.Port)

	return Instance{
		Host:     hostWithPort,
		User:     user,
		Password: pass,
	}, nil
}

func getUserFromEnv(envVars []string) (string, error) {
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 && parts[0] == MINIOUSER {
			return parts[1], nil
		}
	}
	return "", errors.New("MINIO_ROOT_USER environment variable not set")
}

func getPasswordFromEnv(envVars []string) (string, error) {
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 && parts[0] == MINIOPASSWORD {
			return parts[1], nil
		}
	}
	return "", errors.New("MINIO_ROOT_PASSWORD environment variable not set")
}

func getContainerIPAddress(inspect types.ContainerJSON) (string, error) {
	networkMode := string(inspect.HostConfig.NetworkMode)
	networkSettings, found := inspect.NetworkSettings.Networks[networkMode]
	if !found || networkSettings.IPAddress == "" {
		return "", errors.New("container IP address not found")
	}
	return networkSettings.IPAddress, nil
}
