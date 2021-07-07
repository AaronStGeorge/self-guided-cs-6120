package main

import (
	"fmt"
	"os"

	"oooga.ooo/cs-1620/pkg/df"

	"oooga.ooo/cs-1620/pkg/lattice"

	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func defed(instructions []models.Instruction, in lattice.Lattice) lattice.Lattice {
	out := make(utils.Set)
	for _, inst := range instructions {
		if inst.Dest != nil {
			out.Add(*inst.Dest)
		}
	}
	return lattice.UnionMeetSetLattice{utils.Union(in.(lattice.UnionMeetSetLattice).Set, out)}
}

func used(instructions []models.Instruction, in lattice.Lattice) lattice.Lattice {
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
	return lattice.UnionMeetSetLattice{utils.Union(utils.Sub(in.(lattice.UnionMeetSetLattice).Set, defined), used)}
}

func defined(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToInstrs := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToInstrs)

	nameToProgramPoint = make(map[string]*df.ProgramPoint)
	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &df.ProgramPoint{
			Instructions: nameToInstrs[name],
			In:           lattice.UnionMeetSetLattice{Set: make(utils.Set)},
			Out:          lattice.UnionMeetSetLattice{Set: make(utils.Set)},
		}
	}

	workList := []string{namesInOrder[0]}
	df.DF(nameToProgramPoint, cfg, defed, workList, df.Forward)

	return namesInOrder, nameToProgramPoint
}

func live(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToInstrs := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToInstrs)

	nameToProgramPoint = make(map[string]*df.ProgramPoint)
	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &df.ProgramPoint{
			Instructions: nameToInstrs[name],
			In:           lattice.UnionMeetSetLattice{Set: make(utils.Set)},
			Out:          lattice.UnionMeetSetLattice{Set: make(utils.Set)},
		}
	}

	workList := []string{namesInOrder[len(namesInOrder)-1]}
	df.DF(nameToProgramPoint, cfg, used, workList, df.Reverse)

	return namesInOrder, nameToProgramPoint
}

func output(namesInOrder []string, nameToProgramPoint map[string]*df.ProgramPoint) {
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
	var nameToProgramPoint map[string]*df.ProgramPoint
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
