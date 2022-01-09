package utils

import (
	"aaronstgeorge.com/self-guided-cs-1620/pkg/df"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/lattice"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/models"
)

func MakeNameToProgramPoint[T lattice.Lattice[T]](nameToBlock map[string][]models.Instruction, top func() T) map[string]*df.ProgramPoint[T] {
	nameToProgramPoint := make(map[string]*df.ProgramPoint[T])

	for name, block := range nameToBlock {
		nameToProgramPoint[name] = &df.ProgramPoint[T]{
			Instructions: block,
			In:           top(),
			Out:          top(),
		}
	}
	return nameToProgramPoint
}
