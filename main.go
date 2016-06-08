package main

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"github.com/vdemeester/docker-events"
	eventtypes "github.com/docker/engine-api/types/events"
	"os"
	"bufio"
	"io"
)

func main() {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(ctx, options)
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Println(c.Image, "-> ", c.ID)
		reader, pathstat, err := cli.CopyFromContainer(ctx, c.ID, "/usr/bin")
		if err != nil {
			fmt.Println("error from copy:", err)
		} else {
			// open output file
			fo, err := os.Create("output-" + c.ID + ".tar")
			if err != nil {
				panic(err)
			}
			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()

			// make a buffer to keep chunks that are read
			// buffer size==32K
			buf := make([]byte, 32*1024)
			writer := bufio.NewWriter(fo)

			if _, err := io.CopyBuffer(writer, reader, buf); err != nil {
				panic(err)
			}

			fmt.Println("pathstat for copied folder: ", pathstat)
		}

	}


	fmt.Println("test docker events:")

	errChan := events.Monitor(ctx, cli, types.EventsOptions{}, func(event eventtypes.Message) {
		fmt.Printf("%v\n", event)
	})

	if err := <-errChan; err != nil {
		// Do something
	}


	// Call cancel() to get out of the monitor
	defer cancel()
}