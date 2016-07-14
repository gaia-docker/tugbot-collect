package processor

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/gaia-docker/tugbot-collect/log"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"compress/gzip"
	"archive/tar"
	"bufio"
	"net/http"
)

var logger = log.GetLogger("processor")
const resultsTarFile = "results.tar"

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

type results struct {
	testResults         io.Reader
	testResultsPathStat types.ContainerPathStat
	containerInfo       types.ContainerJSON
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

				contResults, err := p.collectResults(ctx, contId)
				if err != nil {
					logger.Error("failed to collect results for container id: ", contId)
					return
				}

				//write the results to disk
				outDirPath, err := writeToDisk(p.outputDir, contId, contResults)
				if err != nil {
					logger.Error("failed to write to disk")
					return
				}

				//send to results Service
				if outDirPath != "" {
					err = p.sendToResultService(outDirPath)
					if err != nil {
						logger.Error("failed to send to results service")
						return
					}
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

func (p Processor) collectResults(ctx context.Context, contId string) (contResults *results, err error) {

	contResults = &results{}
	contResults.containerInfo, err = p.dockerClient.ContainerInspect(ctx, contId)
	if err != nil {
		logger.Error("failed to inspect container with id: ", contId, ", cannot determine results dir location - discard processing this container")
		return nil, err
	}

	//copy the result dir from inside the container into memory
	//The result dir will be return in tar format
	resultDir := contResults.containerInfo.Config.Labels[p.resultsDirLabel]
	logger.Info("going to extract results for container with id: ", contId, ", from this location inside the test container: ", resultDir)
	contResults.testResults, contResults.testResultsPathStat, err = p.dockerClient.CopyFromContainer(ctx, contId, resultDir)
	if err != nil {
		logger.Error("failed to copy result dir from container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return nil, err
	}

	//check that the resultdir is really a directory
	if !contResults.testResultsPathStat.Mode.IsDir() {
		logger.Error("results location:", resultDir, ", inside the container with id: ", contId, ", is not a directory - discard processing of this container")
		return nil, err
	}

	return contResults, nil
}


func writeToDisk(outputDir, contId string, contResults *results) (outDirFullPath string, err error) {

	if outputDir == "/dev/null" {
		logger.Info("output dir is ", outputDir, ", skip writing to disk")
		return  "", nil
	}

	//create output dir (if dir is already exist MkdirAll will return nil)
	outDirFullPath = filepath.Join(outputDir, strings.Replace(contResults.containerInfo.Config.Image, "/", "_", -1) + "-" + contId[:11])
	err = os.MkdirAll(outDirFullPath, 0777)
	if err != nil {
		logger.Error("failed to create output dir: ", outDirFullPath, ", for container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return "", err
	}
	logger.Info("going to write results from container with id: ", contId, ", into this location: ", outDirFullPath)

	// open output file
	resultsTarFullPath := filepath.Join(outDirFullPath, resultsTarFile)
	fo, err := os.Create(resultsTarFullPath)
	if err != nil {
		logger.Error("failed to create tar file on FS for container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return "", err
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

	if _, err := io.CopyBuffer(writer, contResults.testResults, buf); err != nil {
		logger.Error("failed to write to tar file for container with id: ", contId, ", error is: ", err, " - discard processing this container")
		return "", err
	}

	//All is good
	return outDirFullPath, nil
}

func (p Processor) sendToResultService(outDirFullPath string) (err error) {

	if p.resultServiceUrl == "null" || p.resultServiceUrl == "NULL" {
		logger.Info("resultServiceUrl is ", p.resultServiceUrl, ", skip sending results")
		return  nil
	}

	tarGzPath, err := gzipFolder(outDirFullPath)

	f, err := os.Open(tarGzPath)
	if err != nil {
		logger.Error("error openning file: ", err)
		return err
	}
	defer f.Close()

	client := new(http.Client)
	request, err := http.NewRequest("POST", p.resultServiceUrl + "?mainfile=" + resultsTarFile, f)
	request.Header.Add("Content-Type", "application/gzip")
	_, err = client.Do(request)
	if err != nil {
		logger.Error("error uploading file: ", err)
		return err
	}

	logger.Info("sucessfully sent file: ", tarGzPath, ", to results service")
	return nil
}

func gzipFolder(folderPath string) (tarGzPath string, err error) {
	// set up the output file
	tarGzPath = filepath.Join(folderPath, "output.tar.gz")
	file, err := os.Create(tarGzPath)
	if err != nil {
		logger.Error("failed to create: ", tarGzPath, " on disk, error: ", err)
		return "", err
	}
	defer file.Close()

	// set up the gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// grab the paths that need to be added in
	files := []string{
		resultsTarFile,
	}
	// add each file as needed into the current tar archive
	for _,fileName := range files {
		if err := addFileToTar(tw, folderPath, fileName); err != nil {
			logger.Error("failed to add: ", fileName, " to output.tar.gz, error: ", err)
			return "", err
		}
	}

	return tarGzPath, nil
}

func addFileToTar(tw * tar.Writer, folderPath, fileName string) error {
	fullPath := filepath.Join(folderPath, fileName)
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = fileName
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}