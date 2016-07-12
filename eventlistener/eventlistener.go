package eventlistener

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
	"github.com/gaia-docker/tugbot-collect/log"
	"github.com/vdemeester/docker-events"
	"golang.org/x/net/context"
)

var logger = log.GetLogger("eventlistener")

func Register(dockerClient *client.Client, matchlabel string, tasks chan string) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())

		// Setup the event handler on 'die' event (will catch stop and kill and naturally exit
		eventHandler := events.NewHandler(events.ByAction)
		eventHandler.Handle("die", func(event eventtypes.Message) {
			logger.Info("cought event: ", event)
			logger.Info("going to add this container to the tasks list, id: ", event.ID)
			tasks <- event.ID
		})

		//filter only test containers
		f := filters.NewArgs()
		f.Add("label", matchlabel)
		options := types.EventsOptions{Filters: f}

		logger.Info("start monitoring exited test containers with the maching label: ", matchlabel)
		errChan := events.MonitorWithHandler(ctx, dockerClient, options, eventHandler)

		if err := <-errChan; err != nil {
			logger.Error("Event monitor throw this error: ", err)
		}

		defer cancel()
	}()
}
