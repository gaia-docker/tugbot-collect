package processor

import (
	"bufio"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var logger = log.GetLogger("processor")

//Processor is a struct to hold the Tasks channel
//You should use NewProcessor to allocate new one and the Run func to run it
type Processor struct {
	Tasks            chan string
	dockerClient     *client.Client
	outputDir        string
	resultServiceUrl string
	resultsDirLabel  string
	dockerRM         bool
}

//NewProcessor create new Processor and allocates Tasks buffered channel in size 10 to it
func NewProcessor(pDockerClient *client.Client, pOutputDir string, pResultServiceUrl string, pResultsDirLabel string, pDockerRM bool) Processor {
	p := Processor{
		dockerClient:     pDockerClient,
		outputDir:        pOutputDir,
		resultServiceUrl: pResultServiceUrl,
		resultsDirLabel:  pResultsDirLabel,
		dockerRM:         pDockerRM,
	}
	p.Tasks = make(chan string, 10)
	return p
}

//Run the processor (listen to the Tasks channel and performs the collection)
func (p Processor) Run() {
	go func() {
		for task := range p.Tasks {

			//We run each task in a separate goroutine to parallel the extraction
			go func(contId string) {
				logger.Info("processor is going to extract results from container with id: ", contId)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				contInfo, err := p.dockerClient.ContainerInspect(ctx, contId)
				if err != nil {
					logger.Error("failed to inspect container with id: ", contId, ", cannot determine results dir location - discard processing this container")
					return
				}

				//copy the result dir from inside the container into memory
				//The result dir will be return in tar format
				resultDir := contInfo.Config.Labels[p.resultsDirLabel]
				logger.Info("going to extract results for container with id: ", contId, ", from this location inside the test container: ", resultDir)
				reader, pathstat, err := p.dockerClient.CopyFromContainer(ctx, contId, resultDir)
				if err != nil {
					logger.Error("failed to copy result dir from container with id: ", contId, ", error is: ", err, " - discard processing this container")
					return
				}

				//check that the resultdir is really a directory
				if !pathstat.Mode.IsDir() {
					logger.Error("results location:", resultDir, ", inside the container with id: ", contId, ", is not a directory - discard processing of this container")
					return
				}

				//write the results to disk
				writeToDiskStatus := make(chan error)
				writeToDisk(p.outputDir, contInfo.Config.Image, contId, reader, writeToDiskStatus)
				if <-writeToDiskStatus != nil {
					logger.Error("failed to write to disk")
					return
				}

				logger.Info("sucessfully extracted results from container with id: ", contId)

				if p.dockerRM {
					err = p.dockerClient.ContainerRemove(ctx, contId, types.ContainerRemoveOptions{})
					if err != nil {
						logger.Error("failed to remove container with id: ", contId, ", why: ", err)
						return
					}

					logger.Info("sucessfully removed (docker rm) container with id: ", contId)
				}
			}(task)
		}
	}()
}

func writeToDisk(outputDir, imageName, contId string, reader io.ReadCloser, returnVal chan error) {

	go func() {
		if outputDir == "/dev/null" {
			logger.Info("output dir is ", outputDir, ", skip writing to disk")
			returnVal <- nil
			return
		}

		//create output dir (is dir is already exist MkdirAll will return nil
		outDirFullPath := outputDir + string(filepath.Separator) + strings.Replace(imageName, "/", "_", -1) + "-" + contId[:11]
		err := os.MkdirAll(outDirFullPath, 0777)
		if err != nil {
			logger.Error("failed to create output dir: ", outDirFullPath, ", for container with id: ", contId, ", error is: ", err, " - discard processing this container")
			returnVal <- err
			return
		}
		logger.Info("going to write results from container with id: ", contId, ", into this location: ", outDirFullPath)

		// open output file
		fo, err := os.Create(outDirFullPath + string(filepath.Separator) + "results.tar")
		if err != nil {
			logger.Error("failed to create tar file on FS for container with id: ", contId, ", error is: ", err, " - discard processing this container")
			returnVal <- err
			return
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				logger.Error("failed to close FS tar file for container with id: ", contId, ", error is: ", err)
			}
		}()

		// make a buffer to keep chunks that are read
		// buffer size==32K
		buf := make([]byte, 32*1024)
		writer := bufio.NewWriter(fo)

		if _, err := io.CopyBuffer(writer, reader, buf); err != nil {
			logger.Error("failed to write to tar file for container with id: ", contId, ", error is: ", err, " - discard processing this container")
			returnVal <- err
			return
		}

		//All is good, returning nil on the channel
		returnVal <- nil
	}()
}
