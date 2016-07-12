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
	"github.com/gaia-docker/tugbot-collect/eventlistener"
	"github.com/gaia-docker/tugbot-collect/processor"
	"github.com/gaia-docker/tugbot-collect/scanner"
	"os"
	"os/signal"
	"syscall"
)

var logger = log.GetLogger("main")

var dockerrm bool
var scanonstartup bool
var skipevents bool
var outputdir string
var resultserviceurl string
var matchlabel string
var resultsdirlabel string

func main() {

	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "resultserviceurl, u",
			Value:       "http://results-service:8080/results",
			Usage:       "write results to `URL`, if you want to post results to the results service - set this flag with 'null'",
			Destination: &resultserviceurl,
		},
		cli.StringFlag{
			Name:        "outputdir, o",
			Value:       "/tmp/tugbot-collect",
			Usage:       "write results to `DIR_LOCATION`, if you want not to output results - set this flag with the directory '/dev/null'",
			Destination: &outputdir,
		},
		cli.StringFlag{
			Name:        "resultsdirlabel, r",
			Value:       "tugbot.results.dir",
			Usage:       "tugbot-collect will use this label `KEY` to fetch the label value, to find out the results dir of the test container",
			Destination: &resultsdirlabel,
		},
		cli.StringFlag{
			Name:        "matchlabel, m",
			Value:       "tugbot.created.from",
			Usage:       "tugbot-collect will collect results from test containers matching this label `KEY`",
			Destination: &matchlabel,
		},
		cli.BoolFlag{
			Name:        "scanonstartup, e",
			Usage:       "scan for existed containers on startup and extract their results (default is false)",
			Destination: &scanonstartup,
		},
		cli.BoolFlag{
			Name:        "dockerrm, d",
			Usage:       "remove the container after extracting results (default is false)",
			Destination: &dockerrm,
		},
		cli.BoolFlag{
			Name:        "skipevents, s",
			Usage:       "do not register to docker 'die' events (default is false - hence by default we do register to events)",
			Destination: &skipevents,
		},
	}

	app.Name = "tugbot-collect"
	app.Usage = "Collects result from test containers (use TC_LOG_LEVEL env var to change the default which is debug"
	app.Action = start

	if err := app.Run(os.Args); err != nil {
		logger.Error("exiting from main: ", err)
	}

}

func start(c *cli.Context) error {

	logger.Info("tugbot-collect is going to run with this configuration:")
	logger.Info("scanonstartup: ", scanonstartup)
	logger.Info("skipevents: ", skipevents)
	logger.Info("outputdir: ", outputdir)
	logger.Info("dockerrm: ", dockerrm)
	logger.Info("matchlabel: ", matchlabel)
	logger.Info("resultsdirlabel: ", resultsdirlabel)
	logger.Info("resultservice: ", resultserviceurl)

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
	dockerClient, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		logger.Fatal("Failed to create docker client. why: ", err, ". panic the system.")
		panic(err)
	}

	p := processor.NewProcessor(dockerClient, outputdir, resultserviceurl, resultsdirlabel, dockerrm)
	p.Run()

	if scanonstartup {
		scanner.Scan(dockerClient, matchlabel, p.Tasks)
	}

	if !skipevents {
		eventlistener.Register(dockerClient, matchlabel, p.Tasks)
	}

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	logger.Info("awaiting signal")
	<-done
	logger.Info("exiting")

	return nil
}
