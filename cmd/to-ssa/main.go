package main

import (
	"fmt"

	"oooga.ooo/cs-1620/pkg/dominators"
	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func contains(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func existingPhi(block []models.Instruction, v string) bool {
	for _, inst := range block {
		if inst.Op != nil && *inst.Op == "phi" {
			if contains(inst.Args, v) {
				return true
			}
		}
	}
	return false
}

func addPhi(block []models.Instruction, cfg utils.Digraph, v string, t models.Type, df string) []models.Instruction {
	preds := utils.Predecessors(cfg, df)
	args := make([]string, len(preds))
	for i := range args {
		args[i] = v
	}
	phi := "phi"
	newInst := models.Instruction{
		Args:   args,
		Dest:   &v,
		Op:     &phi,
		Labels: preds,
		Type:   &t,
	}

	// TODO: I'm just assuming that a block has a label. They seem to be
	// added places tests, and I think it makes sense that it would be nice
	// if that was just always true. Code to uphold that invariant should be
	// added.

	// Doing an insertion sort here to ensure that phi nodes always have the
	// same order in the block irregardless of the order in which they are
	// added. Ranging over the dominance frontier which is a map and
	// therefore does not have a stable order in go ensures that different
	// paths will be taken and that we need to do this if we want the output
	// to be deterministic.
	i := 1
	for i < len(block) && block[i].Op != nil && *block[i].Op == "phi" && *block[i].Dest < v {
		i++
	}
	out := append(block[:i], append([]models.Instruction{newInst}, block[i:]...)...)
	return out
}

func addPhiNodes(nameToBlock map[string][]models.Instruction, cfg utils.Digraph, nameToDF map[string]utils.Set) {
	// We could tighten up the vars we're searching over some only have one
	// path they are live on.
	vars := utils.NewSet()
	varToType := make(map[string]models.Type)
	for _, block := range nameToBlock {
		for _, inst := range block {
			if inst.Dest != nil {
				vars.Add(*inst.Dest)
				varToType[*inst.Dest] = *inst.Type
			}
		}
	}

	varsToDefiningBlocks := make(map[string]utils.Set)
	for name, block := range nameToBlock {
		for _, inst := range block {
			if inst.Dest != nil && vars.Contains(*inst.Dest) {
				if varsToDefiningBlocks[*inst.Dest] == nil {
					varsToDefiningBlocks[*inst.Dest] = make(utils.Set)
				}
				varsToDefiningBlocks[*inst.Dest].Add(name)
			}
		}
	}

	for variable := range vars {
		worklist := varsToDefiningBlocks[variable]
		for len(worklist) != 0 {
			nextWorkList := make(utils.Set)
			for name := range worklist {
				for df := range nameToDF[name] {
					if !existingPhi(nameToBlock[df], variable) {
						nameToBlock[df] = addPhi(nameToBlock[df], cfg, variable, varToType[variable], df)
						// TODO: there doesn't seem to
						// be a test that will ensure
						// that you do this.
						nextWorkList.Add(df)
					}
				}
			}
			worklist = nextWorkList
		}
	}
}

func rename(name string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph, tree utils.Digraph, vars map[string]*stack, count map[string]int) {
	popsNeeded := make(map[string]int)

	for i, inst := range nameToBlock[name] {
		// Only swap out the args if this is not a phi node. If it is a
		// phi node we will address it specifically below.
		if inst.Op != nil && *inst.Op != "phi" {
			for i, arg := range inst.Args {
				if s, ok := vars[arg]; ok {
					if v, ok := s.peek(); ok {
						inst.Args[i] = v
					} else {
						panic("unreachable")
					}
				}
			}
		}

		if inst.Dest != nil {
			popsNeeded[*inst.Dest] += 1
			v := fmt.Sprintf("%s.%d", *inst.Dest, count[*inst.Dest])
			count[*inst.Dest] += 1
			if s, ok := vars[*inst.Dest]; ok {
				s.push(v)
			} else {
				s := newStack(v)
				vars[*inst.Dest] = s
			}
			inst.Dest = &v
		}
		nameToBlock[name][i] = inst
	}

	for _, successor := range utils.Successors(cfg, name) {
		for _, inst := range nameToBlock[successor] {
			if inst.Op != nil && *inst.Op == "phi" {
				// The location of the label that would be hit
				// when coming from the path we are on
				// determines where we should put the most
				// recent variable.
				idx := firstIndex(inst.Labels, name)
				if idx < 0 {
					panic("unreachable")
				}
				if s, ok := vars[inst.Args[idx]]; ok {
					if v, ok := s.peek(); ok {
						inst.Args[idx] = v
					} else {
						panic("unreachable")
					}
				} else {
					inst.Args[idx] = "__undefined"
				}
			}
		}
	}

	for _, succ := range utils.Successors(tree, name) {
		rename(succ, nameToBlock, cfg, tree, vars, count)
	}

	for name, pops := range popsNeeded {
		if s, ok := vars[name]; ok {
			for ; pops > 0; pops-- {
				if _, ok := s.pop(); !ok {
					panic("unreachable")
				}
			}
			if len(*s) == 0 {
				delete(vars, name)
			}
		} else {
			panic("unreachable")
		}
	}
}

func firstIndex(strings []string, s string) int {
	for i, s1 := range strings {
		if s1 == s {
			return i
		}
	}
	return -1
}

func renameVars(name string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph, tree utils.Digraph) {
	vars := make(map[string]*stack)
	count := make(map[string]int)
	rename(name, nameToBlock, cfg, tree, vars, count)
}

func newStack(strs ...string) *stack {
	s := stack(strs)
	return &s
}

type stack []string

func (s *stack) push(str string) {
	*s = append(*s, str)
}

func (s *stack) peek() (string, bool) {
	if len(*s) == 0 {
		return "", false
	}
	return (*s)[len(*s)-1], true
}

func (s *stack) pop() (string, bool) {
	if len(*s) == 0 {
		return "", false
	}
	out := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return out, true
}

func main() {
	prog := utils.ReadProgram()
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.CFG(namesInOrder, nameToBlock)
	nameToDoms := dominators.Dominators(namesInOrder, nameToBlock, cfg)
	tree := dominators.Tree(namesInOrder, cfg, nameToDoms)
	nameToDF := dominators.Front(namesInOrder, cfg, tree)

	addPhiNodes(nameToBlock, cfg, nameToDF)
	entry := namesInOrder[0]
	renameVars(entry, nameToBlock, cfg, tree)

	namesInOrder, nameToBlock = utils.AddRet(namesInOrder, nameToBlock)
	namesInOrder, nameToBlock = utils.LabelNonEmptyBlocks(namesInOrder, nameToBlock)
	instructions := utils.FlattenBlocks(namesInOrder, nameToBlock)
	prog.Functions[0].Instrs = instructions
	utils.PrintProgram(prog)
}
