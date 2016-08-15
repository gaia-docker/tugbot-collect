package main

import (
	"github.com/docker/engine-api/client"
	"github.com/gaia-docker/tugbot-collect/log"
	"github.com/urfave/cli"
	"github.com/gaia-docker/tugbot-collect/eventlistener"
	"github.com/gaia-docker/tugbot-collect/processor"
	"github.com/gaia-docker/tugbot-collect/scanner"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"errors"
)

var logger = log.GetLogger("main")

var dockerRM bool
var scanOnStartup bool
var skipEvents bool
var outputDir string
var publishTarGzTo string
var publishTestCasesTo string
var matchLabel string
var resultsDirLabel string

func main() {

	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "publishTarGzTo, g",
			Value:       "http://result-service:8080/results",
			Usage:       "send http POST to `URL` with tar.gz payload contains all of the extracted results. To disable - set this flag to 'null'",
			Destination: &publishTarGzTo,
		},
		cli.StringFlag{
			Name:        "publishTestsTo, c",
			Value:       "http://result-service-es:8081/results",
			Usage:       "send http POST to `URL` in json format for any junit test extracted from junit XMLs within the results dir. To disable - set this flag to 'null'",
			Destination: &publishTestCasesTo,
		},
		cli.StringFlag{
			Name:        "outputDir, o",
			Value:       "/tmp/tugbot-collect",
			Usage:       "write results to `DIR_LOCATION`, if you want not to output results - set this flag with the directory '/dev/null'",
			Destination: &outputDir,
		},
		cli.StringFlag{
			Name:        "resultsDirLabel, r",
			Value:       "tugbot.results.dir",
			Usage:       "tugbot-collect will use this label `KEY` to fetch the label value, to find out the results dir of the test container",
			Destination: &resultsDirLabel,
		},
		cli.StringFlag{
			Name:        "matchLabel, m",
			Value:       "tugbot.test",
			Usage:       "tugbot-collect will collect results from test containers matching this label `KEY`",
			Destination: &matchLabel,
		},
		cli.BoolFlag{
			Name:        "scanOnStartup, e",
			Usage:       "scan for existed containers on startup and extract their results (default is false)",
			Destination: &scanOnStartup,
		},
		cli.BoolFlag{
			Name:        "dockerRM, d",
			Usage:       "remove the container after extracting results (default is false)",
			Destination: &dockerRM,
		},
		cli.BoolFlag{
			Name:        "skipEvents, s",
			Usage:       "do not register to docker 'die' events (default is false - by default we register to events and collect results for any stopped or killed test container)",
			Destination: &skipEvents,
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

	if (outputDir == "/dev/null") && (!strings.EqualFold(publishTestCasesTo, "null") || !strings.EqualFold(publishTarGzTo, "null")) {
		return errors.New("outputDir cannot be /dev/null when publishTestCasesTo or publishTarGzTo are in use.")
	}

	logger.Info("tugbot-collect is going to run with this configuration:")
	logger.Info("scanOnStartup: ", scanOnStartup)
	logger.Info("skipEvents: ", skipEvents)
	logger.Info("outputDir: ", outputDir)
	logger.Info("dockerRM: ", dockerRM)
	logger.Info("matchLabel: ", matchLabel)
	logger.Info("resultsDirLabel: ", resultsDirLabel)
	logger.Info("publishTarGzTo: ", publishTarGzTo)
	logger.Info("publishTestCasesTo: ", publishTestCasesTo)

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

	p := processor.NewProcessor(dockerClient, outputDir, publishTarGzTo, publishTestCasesTo, resultsDirLabel, dockerRM)
	p.Run()

	if scanOnStartup {
		scanner.Scan(dockerClient, matchLabel, p.Tasks)
	}

	if !skipEvents {
		eventlistener.Register(dockerClient, matchLabel, p.Tasks)
	}

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	logger.Info("awaiting signal")
	<-done
	logger.Info("exiting")

	return nil
}
