// Trivial dead code elimination

package main

import (
	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func flattenBlocks(namesInOrder []string, nameToBlock map[string][]models.Instruction) []models.Instruction {
	var out []models.Instruction
	for _, name := range namesInOrder {
		out = append(out, nameToBlock[name]...)
	}
	return out
}

// dklPass - a single drop killed local pass
func dklPass(block []models.Instruction) ([]models.Instruction, bool) {
	previous := len(block)

	killedLocal := make(map[int]struct{}) // index of killedLocal stores
	stores := make(map[string]int)
	for i, inst := range block {
		// Uses must be removed before store's are
		// checked. A variable may be an operand to an
		// instruction that will store to the same
		// variable. Doing an increment for example.
		// x++.
		if inst.Args != nil {
			for _, arg := range inst.Args {
				delete(stores, arg)
			}
		}
		if inst.Dest != nil {
			// If this is already in the stores then
			// we know that we have two writes to
			// the same variable without a read.  If
			// this was an increment (for example)
			// of an existing variable the use would
			// have already removed the above if
			// block.
			if i, ok := stores[*inst.Dest]; ok {
				killedLocal[i] = struct{}{}
			}
			stores[*inst.Dest] = i
		}
	}

	var out []models.Instruction
	for i, inst := range block {
		if _, ok := killedLocal[i]; !ok {
			out = append(out, inst)
		}
	}

	return out, len(out) != previous
}

// dkl - drop killed local
// A local optimization that checks for writes that are
// not used before another write.
//
// Example of a dead store:
// 	@main {
// 	  a: int = const 4; <- this can be removed
// 	  a: int = const 2;
// 	  print a;
// 	}
func dkl(function models.Function) (models.Function, bool) {
	namesInOrder, nameToBlock := utils.BasicBlocks(function.Instrs)
	changed := false
	for _, blockName := range namesInOrder {
		// Optimizations that work on an individual
		// block with no control flow are known as
		// "local".
		block := nameToBlock[blockName]
		block, passChanged := dklPass(block)
		nameToBlock[blockName] = block
		changed = changed || passChanged
	}

	function.Instrs = flattenBlocks(namesInOrder, nameToBlock)

	return function, changed
}

// tdce - trivial dead code elimination
func tdce(function models.Function) (models.Function, bool) {
	previous := len(function.Instrs)
	declarations := make(map[string]int)
	for i, inst := range function.Instrs {
		if inst.Dest != nil {
			declarations[*inst.Dest] = i
		}
	}

	for _, inst := range function.Instrs {
		for _, arg := range inst.Args {
			delete(declarations, arg)
		}
	}

	unused := make(map[int]struct{})
	for _, v := range declarations {
		unused[v] = struct{}{}
	}

	var out []models.Instruction
	for i, inst := range function.Instrs {
		if _, ok := unused[i]; !ok {
			out = append(out, inst)
		}
	}
	function.Instrs = out

	return function, len(out) != previous
}

func main() {
	// Optimizations that work on over multiple functions are known as "Inter-procedural"
	prog := utils.ReadProgram()

	passes := []func(models.Function) (models.Function, bool){tdce, dkl}

	// Optimizations that work on an entire function (multiple blocks) are
	// known as "Global".  These will have to deal with control flow between
	// blocks.
	changed := true
	for changed {
		changed = false
		for i, function := range prog.Functions {
			for _, pass := range passes {
				passChanged := false
				function, passChanged = pass(function)
				changed = changed || passChanged
			}
			prog.Functions[i] = function
		}
	}

	utils.PrintProgram(prog)
}
