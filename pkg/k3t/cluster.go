package k3t

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const localhost = "127.0.0.1"

type Cluster struct {
	cli           *client.Client
	image         string
	containerName string
	containerId   string
	port          int
	Logger        zerolog.Logger
}

func (c *Cluster) pullImage(ctx context.Context) error {
	rc, err := c.cli.ImagePull(ctx, c.image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "image pull failed")
	}
	defer rc.Close()
	err = handleImagePull(rc, c.Logger)
	if err != nil {
		return errors.Wrap(err, "error handling image pull event stream")
	}
	return nil
}

func (c *Cluster) createContainer(ctx context.Context) error {
	tcpPort := nat.Port("6443/tcp")
	config := &container.Config{
		Image: c.image,
		Cmd:   []string{"server", "--cluster-init"},
		Env: []string{
			"K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml",
			"K3S_KUBECONFIG_MODE=666",
		},
		ExposedPorts: nat.PortSet{
			tcpPort: struct{}{},
		},
	}
	hostConfig := &container.HostConfig{
		Privileged: true,
		PortBindings: nat.PortMap{
			tcpPort: []nat.PortBinding{
				{
					HostIP:   localhost,
					HostPort: strconv.Itoa(c.port),
				},
			},
		},
	}

	res, err := c.cli.ContainerCreate(
		ctx,
		config, hostConfig, nil, nil,
		c.containerName,
	)

	if err != nil {
		return errors.Wrap(err, "container create failed")
	}
	c.containerId = res.ID

	c.Logger.Info().
		Str("id", res.ID).
		Msg("container created successfully")

	return nil
}

func (c *Cluster) startContainer(ctx context.Context) error {
	err := c.cli.ContainerStart(ctx, c.containerId, types.ContainerStartOptions{})

	if err != nil {
		return errors.Wrap(err, "container start faialed")
	}
	return nil
}

func (c *Cluster) getConfig(ctx context.Context) ([]byte, error) {
	for {
		c.Logger.Trace().
			Msg("waiting for kubeconfig")
		res, err := c.cli.ContainerStatPath(ctx, c.containerId, "/output/kubeconfig.yaml")
		if err != nil {
			time.Sleep(1 * time.Second)
			err = nil
		}
		if res.Size > 0 {
			break
		}
	}
	rr, _, err := c.cli.CopyFromContainer(ctx, c.containerId, "/output/kubeconfig.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "copying kubeconfig from container failed")
	}
	defer rr.Close()
	tr := tar.NewReader(rr)
	hdr, err := tr.Next()
	if err != nil {
		return nil, errors.Wrap(err, "error reading tar header")
	}
	bytes, err := ioutil.ReadAll(tr)
	if err != nil {
		return nil, errors.Wrap(err, "read of kubeconfig failed")
	}
	if _, err := tr.Next(); err != io.EOF {
		return nil, errors.Wrap(err, "expected kubeconfig tar to only contain one file")
	}
	if hdr.Name != "kubeconfig.yaml" {
		return nil, errors.New("expected kubeconfig tar to contain a file named kubeconfig.yaml")
	}

	return bytes, nil
}

func (c *Cluster) waitForPort(ctx context.Context) {
	addr := fmt.Sprintf("%s:%d", localhost, c.port)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Logger.Trace().Msg("Trying to connect to apiserver")
			conn, _ := net.Dial("tcp", addr)
			if conn != nil {
				_ = conn.Close()
				return
			}
		}
	}
}

func (c *Cluster) stopContainer(ctx context.Context) error {
	err := c.cli.ContainerStop(ctx, c.containerId, nil)
	if err != nil {
		return errors.Wrap(err, "container stop failed")
	}
	return nil
}

func (c *Cluster) removeContainer(ctx context.Context) error {
	err := c.cli.ContainerRemove(ctx, c.containerId, types.ContainerRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "container remove failed")
	}
	return nil
}
