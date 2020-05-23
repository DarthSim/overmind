package start

import (
	"fmt"
	"github.com/DarthSim/overmind/v2/utils"
	"net"
)

type portController struct {
	port          int
	step          int
	formationPort int
	formationStep int
	maxTries      int
}

func newPortController(portBase int, portStep int, formationPortStep int, maxTries int) portController {
	return portController{
		port:          portBase,
		step:          portStep,
		formationPort: portBase,
		formationStep: formationPortStep,
		maxTries:      maxTries,
	}
}

func (pc *portController) findNextPort() int {
	nextPort := pc.formationPort
	for tryNum := 0; tryNum < pc.maxTries; tryNum++ {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", nextPort))
		if err == nil {
			listener.Close()
			pc.formationPort = nextPort + pc.formationStep
			return nextPort
		}
		nextPort += pc.formationStep
	}

	utils.Fatal(fmt.Sprintf("Couldn't find available port, tried %d ports starting from %d with step %d.", pc.maxTries, pc.formationPort, pc.formationStep))
	return 0
}

func (pc *portController) nextFormation() {
	pc.port += pc.step
	pc.formationPort = pc.port
}
