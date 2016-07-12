package scanner

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
)

var logger = log.GetLogger("scanner")

func Scan(dockerClient *client.Client, matchlabel string, tasks chan string) {
	go func() {
		logger.Info("scanner trying to find exited containers with the matching label: ", matchlabel)
		ctx, cancel := context.WithCancel(context.Background())
		f := filters.NewArgs()
		f.Add("status", "exited")
		f.Add("label", matchlabel)
		options := types.ContainerListOptions{Filter: f}
		containers, err := dockerClient.ContainerList(ctx, options)
		if err != nil {
			panic(err)
		}

		for _, c := range containers {
			logger.Info("scanner found container: ", c.ID, ", coming from image: ", c.Names[0], ", Sending it to process")
			tasks <- c.ID
		}

		defer cancel()
	}()
}
