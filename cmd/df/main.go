package main

import (
	"fmt"
	"os"

	"aaronstgeorge.com/self-guided-cs-1620/pkg/df"

	"aaronstgeorge.com/self-guided-cs-1620/pkg/lattice"

	dfutils "aaronstgeorge.com/self-guided-cs-1620/pkg/df/utils"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/models"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/utils"
)

func defed(_ string, instructions []models.Instruction, in lattice.UnionMeetSetLattice) lattice.UnionMeetSetLattice {
	out := make(utils.Set)
	for _, inst := range instructions {
		if inst.Dest != nil {
			out.Add(*inst.Dest)
		}
	}
	return lattice.UnionMeetSetLattice{Set: utils.Union(in.Set, out)}
}

func used(_ string, instructions []models.Instruction, in lattice.UnionMeetSetLattice) lattice.UnionMeetSetLattice {
	used := make(utils.Set)
	defined := make(utils.Set)
	for _, inst := range instructions {
		for _, arg := range inst.Args {
			if !defined.Contains(arg) {
				used.Add(arg)
			}
		}
		if inst.Dest != nil {
			defined.Add(*inst.Dest)
		}
	}
	return lattice.UnionMeetSetLattice{Set: utils.Union(utils.Sub(in.Set, defined), used)}
}

func defined(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint[lattice.UnionMeetSetLattice]) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.CFG(namesInOrder, nameToBlock)

	nameToProgramPoint = dfutils.MakeNameToProgramPoint(nameToBlock, func() lattice.UnionMeetSetLattice {
		return lattice.UnionMeetSetLattice{Set: make(utils.Set)}
	})

	workList := []string{namesInOrder[0]}
	df.DF(nameToProgramPoint, cfg, workList, df.Forward, defed)

	return namesInOrder, nameToProgramPoint
}

func live(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint[lattice.UnionMeetSetLattice]) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.CFG(namesInOrder, nameToBlock)

	nameToProgramPoint = dfutils.MakeNameToProgramPoint(nameToBlock, func() lattice.UnionMeetSetLattice {
		return lattice.UnionMeetSetLattice{Set: make(utils.Set)}
	})

	workList := []string{namesInOrder[len(namesInOrder)-1]}
	df.DF(nameToProgramPoint, cfg, workList, df.Reverse, used)

	return namesInOrder, nameToProgramPoint
}

func output[T lattice.Lattice[T]](namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint[T]) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  in:  %s\n", nameToProgramPoint[name].In)
		fmt.Printf("  out: %s\n", nameToProgramPoint[name].Out)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		println("usage: df analysis")
		os.Exit(1)
	}

	prog := utils.ReadProgram()

	var namesInOrder []string
	var nameToProgramPoint map[string]*df.ProgramPoint[lattice.UnionMeetSetLattice]
	switch args[0] {
	case "defined":
		namesInOrder, nameToProgramPoint = defined(prog)
	case "live":
		namesInOrder, nameToProgramPoint = live(prog)
	default:
		println("unknown analysis")
		os.Exit(1)
	}

	output(namesInOrder, nameToProgramPoint)
}
