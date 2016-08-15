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

type RegisterToEvents interface {
	Register()
}

type eventListener struct {
	dockerClient *client.Client
	matchLabel   string
	tasks        chan string
	monitor      func(ctx context.Context, cli client.SystemAPIClient, options types.EventsOptions, handler *events.Handler) chan error
}

func NewEventListener(dockerClient *client.Client, matchLabel string, tasks chan string) RegisterToEvents {
	return &eventListener{
		dockerClient: dockerClient,
		matchLabel:   matchLabel,
		tasks:        tasks,
		monitor:      events.MonitorWithHandler,
	}
}

//Register to "die" events of docker (any container that stopped, killed or exited)
//and writing the container id of any container that has the matchLabel to the tasks channel
func (l *eventListener) Register() {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())

		// Setup the event handler on 'die' event (will catch stop and kill and naturally exit)
		eventHandler := events.NewHandler(events.ByAction)
		eventHandler.Handle("die", func(event eventtypes.Message) {
			logger.Info("cought event: ", event)
			logger.Info("going to add this container to the tasks list, id: ", event.ID)
			l.tasks <- event.ID
		})

		//filter only test containers
		f := filters.NewArgs()
		f.Add("label", l.matchLabel)
		options := types.EventsOptions{Filters: f}

		logger.Info("start monitoring exited test containers with the maching label: ", l.matchLabel)
		errChan := l.monitor(ctx, l.dockerClient, options, eventHandler)

		if err := <-errChan; err != nil {
			logger.Error("Event monitor throw this error: ", err)
		}

		defer cancel()
	}()
}
