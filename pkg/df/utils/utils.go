package utils

import (
	"aaronstgeorge.com/self-guided-cs-1620/pkg/df"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/lattice"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/models"
)

func MakeNameToProgramPoint(nameToBlock map[string][]models.Instruction, top func() lattice.Lattice) map[string]*df.ProgramPoint {
	nameToProgramPoint := make(map[string]*df.ProgramPoint)

	for name, block := range nameToBlock {
		nameToProgramPoint[name] = &df.ProgramPoint{
			Instructions: block,
			In:           top(),
			Out:          top(),
		}
	}
	return nameToProgramPoint
}
