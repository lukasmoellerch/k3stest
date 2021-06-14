package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

const image = "rancher/k3s:v1.21.1-k3s1"

// NewClusterFromEnv creates a new k3s single node cluster that listens to localhost:port once started.
func NewClusterFromEnv(port int) *Cluster {
	containerName := ""

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Logger is disabled by default
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	cluster := &Cluster{
		cli:           cli,
		image:         image,
		containerName: containerName,
		port:          port,
		logger:        logger,
	}

	return cluster
}

// Start creates and starts the container containing the single-node cluster. It returns the certificate that
// can be used to connect to the cluster
func (c *Cluster) Start(ctx context.Context) (*CertificateData, error) {
	var err error
	if err = c.pullImage(ctx); err != nil {
		return nil, err
	}
	if err = c.createContainer(ctx); err != nil {
		return nil, err
	}
	if err = c.startContainer(ctx); err != nil {
		return nil, err
	}
	buf, err := c.getConfig(ctx)
	if err != nil {
		return nil, err
	}
	cert, err := parseKubeconfig(buf)
	if err != nil {
		return nil, err
	}
	c.waitForPort(ctx)
	return &cert, nil
}

// Stop stops the container and removes it.
func (c *Cluster) Stop(ctx context.Context) error {
	var err error
	if err = c.stopContainer(ctx); err != nil {
		return err
	}
	if err = c.removeContainer(ctx); err != nil {
		return err
	}
	return nil
}

func main() {
	bg := context.Background()
	ctx, stop := signal.NotifyContext(bg, os.Interrupt)
	defer stop()

	cluster := NewClusterFromEnv(7443)
	data, err := cluster.Start(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", data)

	<-ctx.Done()
	cluster.logger.Info().
		Msg("shutting down")
	if err := cluster.Stop(bg); err != nil {
		panic(err)
	}
}
