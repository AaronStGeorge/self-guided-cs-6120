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

func dominators(namesInOrder []string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph) (nameToDominators map[string]utils.Set) {
	nameToProgramPoint := dfutils.MakeNameToProgramPoint(nameToBlock, func() lattice.Lattice {
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
			return lattice.IntersetMeetSetLattice{Set: utils.Union(in.(lattice.IntersetMeetSetLattice).Set, this)}
		})

	nameToDominators = make(map[string]utils.Set)
	for name, pp := range nameToProgramPoint {
		nameToDominators[name] = pp.Out.(lattice.IntersetMeetSetLattice).Set
	}

	return nameToDominators
}

type Direction int

const (
	Up Direction = iota
	Down
)

func BFS(cfg utils.Digraph, start string, dir Direction, walk func(name string) bool) bool {
	var nodes []string
	switch dir {
	case Up:
		nodes = utils.Predecessors(cfg, start)
	case Down:
		nodes = utils.Successors(cfg, start)
	}
	for _, name := range nodes {
		if walk(name) {
			return true
		}
	}
	for _, name := range nodes {
		if BFS(cfg, name, dir, walk) {
			return true
		}
	}
	return false
}

func walkUp(cfg utils.Digraph, start string, walk func(name string) bool) {
	BFS(cfg, start, Up, walk)
}

// immediateDom - immediate dominator
func immediateDom(start string, cfg utils.Digraph, dominators utils.Set) (string, bool) {
	s := make(utils.Set)
	s.Add(start)
	dominators = utils.Sub(dominators, s)
	immediateDom := ""
	found := false
	walkUp(cfg, start, func(name string) bool {
		if dominators.Contains(name) {
			immediateDom = name
			found = true
			return true
		}
		return false
	})
	return immediateDom, found
}

func isIDom(dom, sub string, cfg utils.Digraph, nameToDominators map[string]utils.Set) bool {
	if dom == sub {
		return false
	}
	if !nameToDominators[sub].Contains(dom) {
		return false
	}
	idom, ok := immediateDom(sub, cfg, nameToDominators[sub])
	return ok && idom == dom
}

func tree(cfg utils.Digraph, nameToDominators map[string]utils.Set) utils.Digraph {
	out := make(utils.Digraph)
	for dom := range nameToDominators {
		for sub := range nameToDominators {
			if isIDom(dom, sub, cfg, nameToDominators) {
				out[dom] = append(out[dom], sub)
			}
		}
	}
	return out
}

func outputDominators(namesInOrder []string, nameToDominators map[string]utils.Set) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  %s\n", nameToDominators[name])
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		println("usage: dom command")
		os.Exit(1)
	}

	prog := utils.ReadProgram()

	//// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.MakeCFG(namesInOrder, nameToBlock)

	switch args[0] {
	case "dom":
		nameToDominators := dominators(namesInOrder, nameToBlock, cfg)
		outputDominators(namesInOrder, nameToDominators)
	case "tree":
		nameToDominators := dominators(namesInOrder, nameToBlock, cfg)
		utils.OutputDot(namesInOrder, tree(cfg, nameToDominators))
	default:
		println("unknown command")
		os.Exit(1)
	}
}
