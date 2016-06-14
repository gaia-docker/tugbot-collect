package scanner

import (
	"github.com/gaia-docker/tugbot-collect/log"
	"github.com/docker/engine-api/client"
	"golang.org/x/net/context"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

var logger = log.GetLogger("scanner")

func Scan(dockerClient *client.Client, tasks chan string) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		f := filters.NewArgs()
		f.Add("status", "exited")
		options := types.ContainerListOptions{Filter: f}
		containers, err := dockerClient.ContainerList(ctx, options)
		if err != nil {
			panic(err)
		}

		for _, c := range containers {
			logger.Info("scanner found container: ", c.ID, ", comming from image: ", c.Image)
			tasks <- c.ID
		}

		defer cancel()
	}()
}
