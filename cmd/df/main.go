package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

type Lattice interface {
	// Stringer is used for change comparison in the transfer functions
	fmt.Stringer
	// Oh how I pine for generics
	meet(l Lattice) Lattice
}

var _ Lattice = Set{}

type Set map[string]struct{}

func (s Set) String() string {
	if len(s) == 0 {
		return "∅"
	}
	var items []string
	for k := range s {
		items = append(items, k)
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func (s Set) meet(l Lattice) Lattice {
	return union(s, l.(Set))
}

func (s Set) add(item string) {
	s[item] = struct{}{}
}

func union(rhs, lhs Set) Set {
	s := newSet()
	out := s.(Set)
	for item := range rhs {
		out.add(item)
	}
	for item := range lhs {
		out.add(item)
	}
	return out
}

type NewTop func() Lattice

func newSet() Lattice {
	return make(Set)
}

type ProgramPoint struct {
	instructions []models.Instruction
	in           Lattice
	out          Lattice
}

func defed(instructions []models.Instruction, in Lattice) Lattice {
	s := newSet()
	out := s.(Set)
	for _, inst := range instructions {
		if inst.Dest != nil {
			out.add(*inst.Dest)
		}
	}
	return union(in.(Set), out)
}

func defined(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*ProgramPoint) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToInstrs := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToInstrs)

	nameToProgramPoint = make(map[string]*ProgramPoint)
	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &ProgramPoint{instructions: nameToInstrs[name], in: newSet(), out: newSet()}
	}

	workList := []string{namesInOrder[0]}

	for len(workList) != 0 {
		var nextWorkList []string
		for _, name := range workList {
			block := nameToProgramPoint[name]

			for _, pred := range utils.Predecessors(cfg, name) {
				block.in = block.in.meet(nameToProgramPoint[pred].out)
			}

			before := block.out.String()
			block.out = defed(block.instructions, block.in)
			after := block.out.String()

			if before != after {
				nextWorkList = append(workList, utils.Successors(cfg, name)...)
			}
		}
		workList = nextWorkList
	}

	return namesInOrder, nameToProgramPoint
}

func output(namesInOrder []string, nameToProgramPoint map[string]*ProgramPoint) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  in:  %s\n", nameToProgramPoint[name].in)
		fmt.Printf("  out: %s\n", nameToProgramPoint[name].out)
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
	var nameToProgramPoint map[string]*ProgramPoint
	switch args[0] {
	case "defined":
		namesInOrder, nameToProgramPoint = defined(prog)
	default:
		println("unknown analysis")
		os.Exit(1)
	}

	output(namesInOrder, nameToProgramPoint)
}
