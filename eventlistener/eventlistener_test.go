package eventlistener

import (
	"testing"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
	"fmt"
	"time"
	"github.com/docker/docker/api/types/events"
)

//We simulate docker die event and expect to get the die container id in the tasks channel
func TestEventListener(t *testing.T) {

	const labelToMonitor = "tugbot-test"
	tsk := make(chan string, 10)

	//l := NewEventListener(dockerClientMock{}, labelToMonitor, tsk)
	Register(dockerClientMock{}, labelToMonitor, tsk)
	select {
	case res := <-tsk:
		fmt.Println("we recieved the die container id via the tasks chan: ", res)
	case <-time.After(time.Second * 5):
		t.Error("we did not recieved the die container id on the tasks chan after 5 sec!")
	}
}

type dockerClientMock struct {}

func (d dockerClientMock) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	eventsChan := make(chan events.Message, 10)
	errChan := make(chan error, 10)
	event := events.Message{ Type: "container", Action: "die", }
	eventsChan <- event
	return eventsChan, errChan
}
func (d dockerClientMock) Info(ctx context.Context) (types.Info, error) {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) RegistryLogin(ctx context.Context, auth types.AuthConfig) (types.AuthResponse, error) {
	panic("This function not suppose to be called")
}

func (d dockerClientMock) DiskUsage(ctx context.Context) (types.DiskUsage, error) {
	panic("This function not suppose to be called")
}

func (d dockerClientMock) Ping(ctx context.Context) (bool, error) {
	panic("This function not suppose to be called")
}
