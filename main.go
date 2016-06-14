package main

import (
	//"bufio"
	//"fmt"
	"github.com/docker/engine-api/client"
	//"github.com/docker/engine-api/types"
	//eventtypes "github.com/docker/engine-api/types/events"
	"github.com/gaia-docker/tugbot-collect/log"
	"github.com/urfave/cli"
	//"github.com/vdemeester/docker-events"
	//"golang.org/x/net/context"
	//"io"
	"os"
	"os/signal"
	"syscall"
	"github.com/gaia-docker/tugbot-collect/processor"
	"github.com/gaia-docker/tugbot-collect/scanner"
)

var logger = log.GetLogger("main")

func main() {

	var dockerrm bool
	var scanonstartup bool
	var skipevents bool
	var outputdir string

	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "outputdir, o",
			Value:       "/var/logs/tugbot-collect",
			Usage:       "write results to `DIR_LOCATION`, if you want not to output results - set this flag with the directory '/dev/null'",
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
			Usage:       "do not register to docker 'die' events",
			Destination: &skipevents,
		},
	}

	app.Name = "tugbot-collect"
	app.Usage = "Collects result from test containers"
	app.Action = func(c *cli.Context) error {
		logger.Info("tugbot-collect is going to work with these flags - dockerrm: ", dockerrm, ",scanonstartup: ", scanonstartup, ",skipevents: ", skipevents, ",outputdir: ", outputdir)
		return nil
	}

	app.Run(os.Args)

	// Go signal notification works by sending `os.Signal`
	// values on a channel. We'll create a channel to
	// receive these notifications (we'll also make one to
	// notify us when the program can exit).
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		var logger = log.GetLogger("signalgoroutine")
		sig := <-sigs
		logger.Info("got signal: \"", sig, "\", on goroutine. Going to notify main")
		done <- true
	}()

	//Creating docker client
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		logger.Fatal("Failed to create docker client. why: ", err, ". panic the system.")
		panic(err)
	}

	p := processor.NewProcessor(outputdir, dockerrm)
	p.Run()

	if scanonstartup {
		scanner.Scan(cli, p.Tasks)
	}

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	logger.Info("awaiting signal")
	<-done
	logger.Info("exiting")



	/*ctx, cancel := context.WithCancel(context.Background())
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
	defer cancel() */
}
