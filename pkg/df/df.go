package df

import (
	"aaronstgeorge.com/self-guided-cs-1620/pkg/lattice"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/models"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/utils"
)

type Direction int

const (
	Forward Direction = iota
	Reverse
)

type ProgramPoint[T lattice.Lattice[T]] struct {
	Instructions []models.Instruction
	In           T
	Out          T
}

func DF[T lattice.Lattice[T]](nameToProgramPoint map[string]*ProgramPoint[T],
	cfg utils.Digraph,
	initialWorkList []string,
	direction Direction,
	transfer func(name string, instructions []models.Instruction, in T) T) {

	workList := initialWorkList
	for len(workList) != 0 {
		var nextWorkList []string
		for _, name := range workList {
			pp := nameToProgramPoint[name]

			switch direction {
			case Forward:
				for _, pred := range utils.Predecessors(cfg, name) {
					pp.In = pp.In.Meet(nameToProgramPoint[pred].Out)
				}

				before := pp.Out.String()
				pp.Out = transfer(name, pp.Instructions, pp.In)
				after := pp.Out.String()

				if before != after {
					nextWorkList = append(nextWorkList, utils.Successors(cfg, name)...)
				}
			case Reverse:
				for _, pred := range utils.Successors(cfg, name) {
					pp.Out = pp.Out.Meet(nameToProgramPoint[pred].In)
				}

				before := pp.In.String()
				pp.In = transfer(name, pp.Instructions, pp.Out)
				after := pp.In.String()

				if before != after {
					nextWorkList = append(nextWorkList, utils.Predecessors(cfg, name)...)
				}
			}
		}
		workList = nextWorkList
	}
}
