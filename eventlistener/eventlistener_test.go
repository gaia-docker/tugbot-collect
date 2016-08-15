package eventlistener

import (
	"testing"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/vdemeester/docker-events"
	"golang.org/x/net/context"
)

func TestMonitoringOnlyTestContainers(t *testing.T) {

	const labelToMonitor = "tugbot.test"
	testPassed := make(chan struct{})

	l := &eventListener{
		dockerClient: nil,
		matchLabel:   labelToMonitor,
		tasks:        nil,
		monitor:      func(ctx context.Context, cli client.SystemAPIClient, options types.EventsOptions, handler *events.Handler) chan error {
			filterLabel := options.Filters.Get("label")[0]
			if filterLabel != labelToMonitor {
				t.Error("we supposed to filter container with the label: ", labelToMonitor, ", but the filter contained the label: ", filterLabel)
			}

			testPassed<-struct{}{}
			return make(chan error)
		},
		}

	l.Register()
	<-testPassed
}
