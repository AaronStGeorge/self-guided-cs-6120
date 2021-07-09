package main

import (
	"fmt"
	"os"

	"oooga.ooo/cs-1620/pkg/df"
	dfutils "oooga.ooo/cs-1620/pkg/df/utils"
	"oooga.ooo/cs-1620/pkg/lattice"

	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func dom(prog models.Program) {
	//// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToBlock)

	nameToProgramPoint := dfutils.MakeNameToProgramPoint(namesInOrder, nameToBlock, func() lattice.Lattice {
		set := make(utils.Set)
		set.Add(namesInOrder...)
		return lattice.IntersetMeetSetLattice{Set: set}
	})
	// Apply boundary condition
	nameToProgramPoint[namesInOrder[0]].In = lattice.IntersetMeetSetLattice{Set: make(utils.Set)}

	workList := []string{namesInOrder[0]}
	df.DF(nameToProgramPoint,
		cfg,
		workList,
		df.Forward,
		func(name string, _ []models.Instruction, in lattice.Lattice) lattice.Lattice {
			this := make(utils.Set)
			this.Add(name)
			out := lattice.IntersetMeetSetLattice{Set: utils.Union(in.(lattice.IntersetMeetSetLattice).Set, this)}
			return out
		})

	output(namesInOrder, nameToProgramPoint)
}

func output(namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  %s\n", nameToProgramPoint[name].Out)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		println("usage: dom command")
		os.Exit(1)
	}

	prog := utils.ReadProgram()

	switch args[0] {
	case "dom":
		dom(prog)
	default:
		println("unknown command")
		os.Exit(1)
	}
}
