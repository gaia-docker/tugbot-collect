package scanner

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
)

var logger = log.GetLogger("scanner")

//Scan for any exited containers (docker ps -f "status=exited") that has the matchLabel
//and write the container id to the tasks channel
func Scan(dockerClient client.ContainerAPIClient, matchLabel string, tasks chan string) {
	go func() {
		logger.Info("scanner trying to find exited containers with the matching label: ", matchLabel)
		ctx, cancel := context.WithCancel(context.Background())
		f := filters.NewArgs()
		f.Add("status", "exited")
		f.Add("label", matchLabel)
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
