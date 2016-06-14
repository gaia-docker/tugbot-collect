package processor

import "github.com/gaia-docker/tugbot-collect/log"

var logger = log.GetLogger("processor")

type Processor struct {
	Tasks     chan string
	outputdir string
	dockerrm  bool
}

func NewProcessor(p_outputdir string, p_dockerrm bool) Processor {
	p := Processor{
		outputdir: p_outputdir,
		dockerrm:  p_dockerrm,
	}
	p.Tasks = make(chan string, 10)
	return p
}

func (p Processor) Run() {
	go func() {
		for task := range p.Tasks {
			logger.Info("processor recieved container id: ", task)
		}
	}()
}
