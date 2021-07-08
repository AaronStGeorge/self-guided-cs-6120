package utils

import (
	"oooga.ooo/cs-1620/pkg/df"
	"oooga.ooo/cs-1620/pkg/lattice"
	"oooga.ooo/cs-1620/pkg/models"
)

func MakeNameToProgramPoint(namesInOrder []string, nameToBlock map[string][]models.Instruction, top func() lattice.Lattice) map[string]*df.ProgramPoint {
	nameToProgramPoint := make(map[string]*df.ProgramPoint)

	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &df.ProgramPoint{
			Instructions: nameToBlock[name],
			In:           top(),
			Out:          top(),
		}
	}
	return nameToProgramPoint
}
