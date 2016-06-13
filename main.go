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
	"github.com/urfave/cli"
)

func main() {
	var dockerrm bool
	var scanonstartup bool
	var skipevents bool
	var outputdir string

	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name:        "outputdir, o",
			Value:       "/var/logs/tugbot-collect",
			Usage:       "write results to `DIR_LOCATION`, if you want not to output results set this flag with the directory '/dev/null'",
			Destination: &outputdir,
		},
		cli.BoolFlag{
			Name:        "scanonstartup, e",
			Usage:       "scan for existed containers on startup and extract their results",
			Destination: &scanonstartup,
		},
		cli.BoolFlag{
			Name:        "dockerrm, d",
			Usage:       "remove the container after extracting results",
			Destination: &dockerrm,
		},
		cli.BoolFlag{
			Name:        "skipevents, s",
			Usage:       "do not register to docker 'die' event",
			Destination: &skipevents,
		},
	}

	app.Name = "tugbot-collect"
	app.Usage = "Collects result from test containers"
	app.Action = func(c *cli.Context) error {
		fmt.Println("tugbot-collect is going to work with these flags - dockerrm: ", dockerrm, ",scanonstartup: ", scanonstartup, ",skipevents: ", skipevents,",outputdir: ", outputdir)
		return nil
	}

	app.Run(os.Args)

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