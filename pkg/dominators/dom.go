package dominators

import (
	"oooga.ooo/cs-1620/pkg/df"
	dfutils "oooga.ooo/cs-1620/pkg/df/utils"
	"oooga.ooo/cs-1620/pkg/lattice"
	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func Front(namesInOrder []string, cfg utils.Digraph, domTree utils.Digraph) map[string]utils.Set {
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

func Tree(cfg utils.Digraph, nameToDominators map[string]utils.Set) utils.Digraph {
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

func Dominators(namesInOrder []string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph) (nameToDominators map[string]utils.Set) {
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

// immediateDominatorFromDoms - immediate dominator calculated from the dominators set
func immediateDominatorFromDoms(start string, cfg utils.Digraph, dominators utils.Set) (string, bool) {
	s := make(utils.Set)
	s.Add(start)
	dominators = utils.Sub(dominators, s)
	immediateDom := ""
	found := false
	utils.WalkUp(cfg, start, func(name string) bool {
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
