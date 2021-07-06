package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

type Direction int

const (
	Forward Direction = iota
	Reverse
)

type Lattice interface {
	// Stringer is used for change comparison in the transfer functions
	fmt.Stringer
	// Oh how I pine for generics
	meet(l Lattice) Lattice
}

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

func (s Set) add(item string) {
	s[item] = struct{}{}
}

func (s Set) contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s Set) remove(item string) {
	delete(s, item)
}

var _ Lattice = UnionMeetSetLattice{}

type UnionMeetSetLattice struct {
	Set
}

func (s UnionMeetSetLattice) meet(l Lattice) Lattice {
	return UnionMeetSetLattice{union(s.Set, l.(UnionMeetSetLattice).Set)}
}

func union(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.add(item)
	}
	for item := range lhs {
		out.add(item)
	}
	return out
}

func sub(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.add(item)
	}
	for item := range lhs {
		out.remove(item)
	}
	return out
}

type ProgramPoint struct {
	instructions []models.Instruction
	in           Lattice
	out          Lattice
}

func defed(instructions []models.Instruction, in Lattice) Lattice {
	out := make(Set)
	for _, inst := range instructions {
		if inst.Dest != nil {
			out.add(*inst.Dest)
		}
	}
	return UnionMeetSetLattice{union(in.(UnionMeetSetLattice).Set, out)}
}

func used(instructions []models.Instruction, in Lattice) Lattice {
	used := make(Set)
	defined := make(Set)
	for _, inst := range instructions {
		for _, arg := range inst.Args {
			if !defined.contains(arg) {
				used.add(arg)
			}
		}
		if inst.Dest != nil {
			defined.add(*inst.Dest)
		}
	}
	return UnionMeetSetLattice{union(sub(in.(UnionMeetSetLattice).Set, defined), used)}
}

func df(nameToProgramPoint map[string]*ProgramPoint,
	cfg utils.Digraph,
	transfer func(instructions []models.Instruction, in Lattice) Lattice,
	initialWorkList []string,
	direction Direction) {

	workList := initialWorkList
	for len(workList) != 0 {
		var nextWorkList []string
		for _, name := range workList {
			pp := nameToProgramPoint[name]

			switch direction {
			case Forward:
				for _, pred := range utils.Predecessors(cfg, name) {
					pp.in = pp.in.meet(nameToProgramPoint[pred].out)
				}

				before := pp.out.String()
				pp.out = transfer(pp.instructions, pp.in)
				after := pp.out.String()

				if before != after {
					nextWorkList = append(workList, utils.Successors(cfg, name)...)
				}
			case Reverse:
				for _, pred := range utils.Successors(cfg, name) {
					pp.out = pp.out.meet(nameToProgramPoint[pred].in)
				}

				before := pp.in.String()
				pp.in = transfer(pp.instructions, pp.out)
				after := pp.in.String()

				if before != after {
					nextWorkList = append(workList, utils.Predecessors(cfg, name)...)
				}
			}
		}
		workList = nextWorkList
	}
}

func defined(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*ProgramPoint) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToInstrs := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToInstrs)

	nameToProgramPoint = make(map[string]*ProgramPoint)
	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &ProgramPoint{instructions: nameToInstrs[name],
			in:  UnionMeetSetLattice{make(Set)},
			out: UnionMeetSetLattice{make(Set)}}
	}

	workList := []string{namesInOrder[0]}
	df(nameToProgramPoint, cfg, defed, workList, Forward)

	return namesInOrder, nameToProgramPoint
}

func live(prog models.Program) (namesInOrder []string, nameToProgramPoint map[string]*ProgramPoint) {
	// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToInstrs := utils.BasicBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToInstrs)

	nameToProgramPoint = make(map[string]*ProgramPoint)
	for _, name := range namesInOrder {
		nameToProgramPoint[name] = &ProgramPoint{instructions: nameToInstrs[name],
			in:  UnionMeetSetLattice{make(Set)},
			out: UnionMeetSetLattice{make(Set)}}
	}

	workList := []string{namesInOrder[len(namesInOrder)-1]}
	df(nameToProgramPoint, cfg, used, workList, Reverse)

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
	case "live":
		namesInOrder, nameToProgramPoint = live(prog)
	default:
		println("unknown analysis")
		os.Exit(1)
	}

	output(namesInOrder, nameToProgramPoint)
}
