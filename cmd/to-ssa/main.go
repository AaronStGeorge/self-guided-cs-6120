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

	out := []models.Instruction{block[0], newInst}
	out = append(out, block[1:]...)
	return out
}

func addPhiNodes(nameToBlock map[string][]models.Instruction, cfg utils.Digraph, nameToDF map[string]utils.Set) {
	// We could tighten up the vars we're searching over
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

	varsToDefiningBlocks := make(map[string][]string)
	for name, block := range nameToBlock {
		for _, inst := range block {
			if inst.Dest != nil && vars.Contains(*inst.Dest) {
				varsToDefiningBlocks[*inst.Dest] = append(varsToDefiningBlocks[*inst.Dest], name)
			}
		}
	}

	for variable := range vars {
		worklist := varsToDefiningBlocks[variable]
		for len(worklist) != 0 {
			var nextWorkList []string
			for _, name := range worklist {
				for df := range nameToDF[name] {
					if !existingPhi(nameToBlock[df], variable) {
						nameToBlock[df] = addPhi(nameToBlock[df], cfg, variable, varToType[variable], df)
						// TODO: there doesn't seem to
						// be a test that will ensure
						// that you do this.
						nextWorkList = append(nextWorkList, df)
					}
				}
			}
			worklist = nextWorkList
		}
	}
}

func rename(name string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph, tree utils.Digraph, vars map[string]*ssaVar) {
	popsNeeded := make(map[string]int)

	for i, inst := range nameToBlock[name] {
		// Only swap out the args if this is not a phi node. If it is the other thing will get it
		if inst.Op != nil && *inst.Op != "phi" {
			for i, arg := range inst.Args {
				if sv, ok := vars[arg]; ok {
					if v, ok := sv.s.peek(); ok {
						inst.Args[i] = v
					} else {
						panic("unreachable")
					}
				}
			}
		}

		if inst.Dest != nil {
			popsNeeded[*inst.Dest] += 1
			v := ""
			if sv, ok := vars[*inst.Dest]; ok {
				v = fmt.Sprintf("%s.%d", *inst.Dest, sv.c)
				sv.s.push(v)
				sv.c++
			} else {
				v = *inst.Dest + ".0"
				sv := ssaVar{s: newStack(v), c: 1}
				vars[*inst.Dest] = &sv
			}
			inst.Dest = &v
		}
		nameToBlock[name][i] = inst
	}

	successors := utils.Successors(cfg, name)
	for _, successor := range successors {
		for _, inst := range nameToBlock[successor] {
			if inst.Op != nil && *inst.Op == "phi" {

				// TODO: explain this shit
				idx := firstIndex(inst.Labels, name)
				if idx < 0 {
					panic("unreachable")
				}
				if sv, ok := vars[inst.Args[idx]]; ok {
					if v, ok := sv.s.peek(); ok {
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
		rename(succ, nameToBlock, cfg, tree, vars)
	}

	for name, pops := range popsNeeded {
		if sv, ok := vars[name]; ok {
			for ; pops > 0; pops-- {
				if _, ok := sv.s.pop(); !ok {
					panic("unreachable")
				}
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

type ssaVar struct {
	s *stack
	c int
}

func renameVars(name string, nameToBlock map[string][]models.Instruction, cfg utils.Digraph, tree utils.Digraph) {
	vars := make(map[string]*ssaVar)
	rename(name, nameToBlock, cfg, tree, vars)
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

	prog.Functions[0].Instrs = utils.FlattenBlocks(namesInOrder, nameToBlock)
	utils.PrintProgram(prog)
}
