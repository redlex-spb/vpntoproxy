package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"vpntoproxy/internal/config"
)

type Client struct {
	cli *client.Client
	cnf *config.Docker
}

const (
	imagePath = "deployments/docker-vpnwithproxy/docker-vpnwithproxy.tar"
)

// Инициализация модуля «Docker»
func New() (*Client, error) {
	logrus.Debug(">>> Initialization Docker package")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Debug("Error create client docker")
		return nil, err
	}

	logrus.Debug("<<< End of Initialization Docker package")

	return &Client{
		cli: cli,
		cnf: config.Get().Docker,
	}, nil
}

// Метод получения списка контейнеров
func (cl *Client) GetContainersList() (containers []types.Container, err error) {
	logrus.Debug(">>> Starting get containers list")

	containers, err = cl.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		logrus.Debug("Error getting containers list")
		return nil, err
	}

	logrus.Debug("Container list received successfully")
	logrus.Debug("<<< Ending get containers list")

	return containers, nil
}

// Метод получения списка контейнеров с образом vpn
func (cl *Client) ContainersVPNList() (res []types.Container, err error) {
	logrus.Debug(">>> Starting get containers list with image '%s'", cl.cnf.ImageName)

	containers, err := cl.GetContainersList()
	if err != nil {
		return nil, err
	}

	for _, _container := range containers {
		if _container.Image == cl.cnf.ImageName {
			res = append(res, _container)
		}
	}

	logrus.Debug("Container list received successfully")
	logrus.Debug("<<< Ending get containers list with image '%s'", cl.cnf.ImageName)

	return res, nil
}

// Метод создания контейнера
func (cl *Client) RunContainer(config *container.Config, hostConfig *container.HostConfig, basename string) (
	*container.ContainerCreateCreatedBody, error) {

	logrus.Debug(">>> Starting create container")
	logrus.Debug("Container params:", config)
	logrus.Debug("Host params:", hostConfig)

	ctx := context.Background()

	resp, err := cl.cli.ContainerCreate(ctx, config, hostConfig, nil, nil,
		cl.cnf.ServicePrefix+strings.Split(basename, ".")[0])

	if err != nil {
		logrus.Debug("Failed create container")
		return nil, err
	}

	if err := cl.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logrus.Debug("Failed start container")
		return nil, err
	}

	logrus.Debug("Run container succesfully")
	logrus.Debug("<<< Ending create container")

	return &resp, nil
}

// Метод получения списка образов
func (cl *Client) GetListImages() (images []types.ImageSummary, err error) {
	logrus.Debug(">>> Starting get images list")
	ctx := context.Background()

	images, err = cl.cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	logrus.Debug("<<< Ending get images list")

	return images, nil
}

// Метод создания образа
func (cl *Client) BuildImage(tag string, bctx string) (*types.ImageBuildResponse, error) {
	logrus.Debug(">>> Starting build image")
	logrus.Debug("Image tag:", tag)
	logrus.Debug("Build context:", bctx)

	ctx := context.Background()

	options := types.ImageBuildOptions{
		Tags: []string{tag},
	}

	f, err := os.Open(imagePath)
	if err != nil {
		logrus.Debug("Error open dockerfile")
		logrus.Error(err)
		return nil, err
	}

	logrus.Debug("Image options:", options)

	buildResponse, err := cl.cli.ImageBuild(ctx, f, options)

	if err != nil {
		logrus.Debug("Error build image")
		logrus.Error(err)
		return nil, err
	}

	logrus.Debug("Image created succesfully")
	logrus.Debug("<<< Ending build image")

	return &buildResponse, nil
}

// Метод получения контейнера по идентификатору
func (cl *Client) GetContainerByID(ID string) (*types.Container, error) {
	logrus.Debug(">>> Starting get container by ID")
	logrus.Debug("Container ID:", ID)

	_filters := filters.NewArgs()
	_filters.Add("id", ID)

	_container, err := cl.cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: _filters,
	})
	if err != nil {
		logrus.Debug("Container not finded")
		return nil, err
	}

	logrus.Debug("<<< Ending get container by ID")

	if len(_container) > 0 {
		logrus.Debug("Container finded succesfully")
		logrus.Debug(&_container[0])
		return &_container[0], nil
	}

	return nil, fmt.Errorf("Container not finded")
}

// Метод закрытия контейнера
func (cl *Client) Kill(id string) (bool, error) {
	logrus.Debug(">>> Starting kill container")
	logrus.Debug("Container ID:", id)

	err := cl.cli.ContainerKill(context.Background(), id, "SIGKILL")
	if err != nil {
		logrus.Debug("Error, kill container failed")
		return false, err
	}

	logrus.Debug("Container killed succesfully")
	logrus.Debug("<<< Ending kill container")

	return true, nil
}

// Метод удаления контейнера
func (cl *Client) Remove(id string) (bool, error) {
	logrus.Debug(">>> Starting remove container")
	logrus.Debug("Container ID:", id)

	err := cl.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
	if err != nil {
		logrus.Debug("Error, remove container failed")
		return false, err
	}

	logrus.Debug("Container removed succesfully")
	logrus.Debug("<<< Ending remove container")

	return true, nil
}

func (cl *Client) CheckVPN(id string) (bool, error) {
	logrus.Debug(">>> Starting check vpn")
	logrus.Debug("Container ID:", id)

	logs, err := cl.cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		logrus.Debug("Error, get logs container failed")
		return false, err
	}

	logsBytes, err := ioutil.ReadAll(logs)

	logrus.Debug("Vpn checked succesfully")
	logrus.Debug("<<< Ending check vpn")

	return bytes.Contains(logsBytes, []byte("Initialization Sequence Completed")), nil
}
