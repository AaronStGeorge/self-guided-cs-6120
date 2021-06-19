// Trivial dead code elimination

package main

import (
	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/optimisations"
	"oooga.ooo/cs-1620/pkg/utils"
)

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

	passes := []func(models.Function) (models.Function, bool){optimisations.TDCE, optimisations.DKL}

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
