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

// immediateDominatorFromDoms - immediate dominator calculated from the dominators set
func immediateDominatorFromDoms(start string, cfg utils.Digraph, dominators utils.Set) (string, bool) {
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
	idom, ok := immediateDominatorFromDoms(sub, cfg, nameToDominators[sub])
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

// immediateDominatorFromTree - immediate dominators from the dominator tree
func immediateDominatorFromTree(name string, domTree utils.Digraph) (string, bool) {
	preds := utils.Predecessors(domTree, name)
	if len(preds) > 1 {
		panic("assertion error: dom tree malformed")
	}
	if len(preds) == 1 {
		return preds[0], true
	}
	return "", false
}

func front(namesInOrder []string, cfg utils.Digraph, domTree utils.Digraph) map[string]utils.Set {
	immediateDominator := func(node string) string {
		if imDdom, ok := immediateDominatorFromTree(node, domTree); ok {
			return imDdom
		}
		// The only thing that won't have an immediate dominator will be
		// the entry point. While running up the dominators from a
		// predecessor of a join point we should not need to ask what
		// the immediate dominator of the entry point is because that
		// will always be a strict dominator of any join point, so our
		// while loop will short circut.
		panic("assertion error: no immediate dominator")
	}

	// initialization
	nameToFront := make(map[string]utils.Set)
	for _, name := range namesInOrder {
		nameToFront[name] = utils.NewSet()
	}

	// Compute dominance frontier
	// Engineering a Compiler pp. 499
	for _, name := range namesInOrder {
		preds := utils.Predecessors(cfg, name)
		if len(preds) >= 2 {
			joinPoint := name
			for _, pred := range preds {
				runner := pred
				if idom, ok := immediateDominatorFromTree(joinPoint, domTree); ok {
					for runner != idom {
						nameToFront[runner] = utils.Union(nameToFront[runner], utils.NewSet(joinPoint))
						runner = immediateDominator(runner)
					}
				}
			}
		}
	}
	return nameToFront
}

func outputNameToSet(namesInOrder []string, nameToSet map[string]utils.Set) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  %s\n", nameToSet[name])
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
		outputNameToSet(namesInOrder, nameToDominators)
	case "tree":
		nameToDominators := dominators(namesInOrder, nameToBlock, cfg)
		utils.OutputDot(namesInOrder, tree(cfg, nameToDominators))
	case "front":
		nameToDominators := dominators(namesInOrder, nameToBlock, cfg)
		domTree := tree(cfg, nameToDominators)
		front := front(namesInOrder, cfg, domTree)
		outputNameToSet(namesInOrder, front)
	default:
		println("unknown command")
		os.Exit(1)
	}
}
