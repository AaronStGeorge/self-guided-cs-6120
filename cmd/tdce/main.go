// Trivial dead code elimination

package main

import (
	"aaronstgeorge.com/self-guided-cs-1620/pkg/models"
	"aaronstgeorge.com/self-guided-cs-1620/pkg/utils"
)

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

// DKL - drop killed local
// A local optimization that checks for writes that are
// not used before another write.
//
// Example of a dead store:
// 	@main {
// 	  a: int = const 4; <- this can be removed
// 	  a: int = const 2;
// 	  print a;
// 	}
func DKL(function models.Function) (models.Function, bool) {
	namesInOrder, nameToBlock := utils.BasicBlocks(function)
	changed := false
	for _, blockName := range namesInOrder {
		block := nameToBlock[blockName]
		block, passChanged := dklPass(block)
		nameToBlock[blockName] = block
		changed = changed || passChanged
	}

	function.Instrs = utils.FlattenBlocks(namesInOrder, nameToBlock)

	return function, changed
}

// TDCE - trivial dead code elimination
func TDCE(function models.Function) (models.Function, bool) {
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
	// If a function works over:
	//
	//  multiple functions it is known as "inter-procedural".
	//
	//  an entire function (potentially multiple blocks) it is known as
	//  "global".
	//
	//  an individual block (meaning no control flow) it is known as
	//  "local".
	prog := utils.ReadProgram()

	passes := []func(models.Function) (models.Function, bool){TDCE, DKL}

	changed := true
	for changed {
		changed = false
		for i, function := range prog.Functions {
			for _, pass := range passes {
				passChanged := false
				function, passChanged = pass(function)
				changed = changed || passChanged
			}
			// The for loop uses value semantics, we aren't
			// modifying the real function. To replace it we have to
			// index into the functions array.
			prog.Functions[i] = function
		}
	}

	utils.PrintProgram(prog)
}
