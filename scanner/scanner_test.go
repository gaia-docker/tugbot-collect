package scanner

import (
	"testing"
	"golang.org/x/net/context"
	"github.com/docker/docker/api/types"
	"fmt"
	"time"
	"io"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"errors"
)

const labelToScan = "tugbot-test"

func TestScanner(t *testing.T) {

	tsk := make(chan string, 10)
	Scan(dockerClientMock{}, labelToScan, tsk)

	select {
	case res := <-tsk:
		fmt.Println("we recieved the die container id via the tasks chan: ", res)
	case <-time.After(time.Second * 5):
		t.Error("we did not recieved the die container id on the tasks chan after 5 sec!")
	}
}

type dockerClientMock struct {}

func (d dockerClientMock) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if options.Filter.Get("status")[0] == "exited" && options.Filter.Get("label")[0] == labelToScan {
		conts := []types.Container { { ID: "123456", Names: []string {"MockContainer"} },
		}
		return conts, nil
	}

	return nil, errors.New("failure in test: status is not exited or/and matchlabel is not equal")
}

func (d dockerClientMock) ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerCommit(ctx context.Context, container string, options types.ContainerCommitOptions) (types.ContainerCommitResponse, error) {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (types.ContainerCreateResponse, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerDiff(ctx context.Context, container string) ([]types.ContainerChange, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.ContainerExecCreateResponse, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExecResize(ctx context.Context, execID string, options types.ResizeOptions) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerExport(ctx context.Context, container string) (io.ReadCloser, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerInspectWithRaw(ctx context.Context, container string, getSize bool) (types.ContainerJSON, []byte, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerKill(ctx context.Context, container, signal string) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerPause(ctx context.Context, container string) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock)ContainerRemove(ctx context.Context, container string, options types.ContainerRemoveOptions) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerRename(ctx context.Context, container, newContainerName string) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerResize(ctx context.Context, container string, options types.ResizeOptions) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerRestart(ctx context.Context, container string, timeout *time.Duration) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerStatPath(ctx context.Context, container, path string) (types.ContainerPathStat, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerStats(ctx context.Context, container string, stream bool) (types.ContainerStats, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerStop(ctx context.Context, container string, timeout *time.Duration) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerTop(ctx context.Context, container string, arguments []string) (types.ContainerProcessList, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerUnpause(ctx context.Context, container string) error  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock)ContainerUpdate(ctx context.Context, container string, updateConfig container.UpdateConfig) (types.ContainerUpdateResponse, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) ContainerWait(ctx context.Context, container string) (int, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) CopyFromContainer(ctx context.Context, container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)  {
	panic("This function not suppose to be called")
}
func (d dockerClientMock) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error  {
	panic("This function not suppose to be called")
}

func (d dockerClientMock) ContainersPrune(ctx context.Context, cfg types.ContainersPruneConfig) (types.ContainersPruneReport, error) {
	panic("This function not suppose to be called")
}
