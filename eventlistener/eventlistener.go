package eventlistener

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
)

var logger = log.GetLogger("eventlistener")

//Register to "die" events of docker (any container that stopped, killed or exited)
//and writing the container id of any container that has the matchLabel to the tasks channel
func Register(dockerClient client.SystemAPIClient, matchLabel string, tasks chan string) {
	go func() {
		//filter only "die" test containers
		f := filters.NewArgs()
		f.Add("label", matchLabel)
		f.Add("event", "die")
		f.Add("type", "container")
		options := types.EventsOptions{Filters: f}

		//Listen to events
		ctx, cancel := context.WithCancel(context.Background())
		eventsChan, errChan := dockerClient.Events(ctx, options)
		logger.Info("start monitoring exited test containers with the maching label: ", matchLabel)

		go func(){
			for event := range eventsChan {
				logger.Info("cought event: ", event)
				logger.Info("going to add this container to the tasks list, id: ", event.ID)
				tasks <- event.ID
			}

		}()

		if err := <-errChan; err != nil {
			logger.Error("Event monitor throw this error: ", err)
		}

		defer cancel()
	}()
}
