package eventlistener

import (
	"testing"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"io"
	"fmt"
	"bytes"
	"time"
)

//We simulate docker die event and expect to get the die container id in the tasks channel
func TestEventListener(t *testing.T) {

	const labelToMonitor = "tugbot.test"
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

func (d dockerClientMock) Events(ctx context.Context, options types.EventsOptions) (io.ReadCloser, error) {
	//A sample event was extracted using this code:
	//f.Add("event", "die")
	//options := types.EventsOptions{Filters: f}
	//b, _ := dockerClient.Events(ctx, options)
	//arr := make([]byte, 1024)
	//b.Read(arr)
	//fmt.Println(string(arr))
	return nopCloser{bytes.NewBufferString("{\"status\":\"die\",\"id\":\"5fbe2d71593def3b1ac43fbdebfaa42528fa5f857415bd25af64bf61aed22b79\",\"from\":\"hello-world\",\"Type\":\"container\",\"Action\":\"die\",\"Actor\":{\"ID\":\"5fbe2d71593def3b1ac43fbdebfaa42528fa5f857415bd25af64bf61aed22b79\",\"Attributes\":{\"exitCode\":\"0\",\"image\":\"hello-world\",\"name\":\"serene_borg\"}},\"time\":1471266369,\"timeNano\":1471266369413735698}")}, nil
}
func (d dockerClientMock) Info(ctx context.Context) (types.Info, error) {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) RegistryLogin(ctx context.Context, auth types.AuthConfig) (types.AuthResponse, error) {
	panic("This function not suppose to be called")
}

//struct to fit the io.ReadCloser interface
type nopCloser struct {
	io.Reader
}
func (nopCloser) Close() error {
	return nil
}


