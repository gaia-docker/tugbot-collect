package processor

import (
	"bufio"
	"github.com/docker/engine-api/client"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/docker/engine-api/types"
)

var logger = log.GetLogger("processor")

type Processor struct {
	Tasks           chan string
	dockerClient    *client.Client
	outputdir       string
	resultsdirlabel string
	dockerrm        bool
}

func NewProcessor(p_dockerClient *client.Client, p_outputdir string, p_resultsdirlabel string, p_dockerrm bool) Processor {
	p := Processor{
		dockerClient:    p_dockerClient,
		outputdir:       p_outputdir,
		resultsdirlabel: p_resultsdirlabel,
		dockerrm:        p_dockerrm,
	}
	p.Tasks = make(chan string, 10)
	return p
}

func (p Processor) Run() {
	go func() {
		for task := range p.Tasks {

			//We run each task in a separate goroutine to parallel the extraction
			go func(contId string) {
				logger.Info("processor is going to extract results from container with id: ", contId)

				ctx, cancel := context.WithCancel(context.Background())
				contInfo, err := p.dockerClient.ContainerInspect(ctx, contId)
				if err != nil {
					logger.Error("failed to inspect container with id: ", contId, ", cannot determine results dir location - discard processing this container")
					return
				}

				//copy the result dir from inside the container into memory
				//The result dir will be return in tar format
				resultDir := contInfo.Config.Labels[p.resultsdirlabel]
				logger.Info("going to extract results for container with id: ", contId, ", from this location inside the test container: ", resultDir)
				reader, pathstat, err := p.dockerClient.CopyFromContainer(ctx, contId, resultDir)
				if err != nil {
					logger.Error("failed to copy result dir from container with id: ", contId, ", error is: ", err, " - discard processing this container")
					return
				}

				//check that the resultdir is really a directory
				if !pathstat.Mode.IsDir() {
					logger.Error("results location:", resultDir, ", inside the container with id: ", contId, ", is not a directory - discard processing this container")
					return
				}

				//write the results to disk
				err = writeToDisk(p.outputdir, contInfo.Config.Image, contId, reader)
				if err != nil {
					logger.Error("failed to write to disk")
					return
				}

				logger.Info("sucessfully extracted results from container with id: ", contId)

				if p.dockerrm {
					err = p.dockerClient.ContainerRemove(ctx, contId, types.ContainerRemoveOptions{})
					if err != nil {
						logger.Error("failed to remove container with id: ", contId, ", why: ", err)
						return
					}

					logger.Info("sucessfully removed (docker rm) container with id: ", contId)
				}
				defer cancel()
			}(task)
		}
	}()
}

func writeToDisk(outputDir, imageName, contId string, reader io.ReadCloser) error {

	if outputDir == "/dev/null" {
		logger.Info("output dir is ", outputDir, ", skip writing to disk")
		return nil
	}

	//create output dir (is dir is already exist MkdirAll will return nil
	outDirFullPath := outputDir + string(filepath.Separator) + strings.Replace(imageName, "/", "_", -1) + "-" + contId[:11]
	err := os.MkdirAll(outDirFullPath, 0777)
	if err != nil {
		logger.Error("failed to create output dir: ", outDirFullPath, ", for container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return err
	}
	logger.Info("going to write results from container with id: ", contId, ", into this location: ", outDirFullPath)

	// open output file
	fo, err := os.Create(outDirFullPath + string(filepath.Separator) + "results.tar")
	if err != nil {
		logger.Error("failed to create tar file on FS for container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return err
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
		return err
	}

	return nil
}
