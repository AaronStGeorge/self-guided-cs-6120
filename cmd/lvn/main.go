// Local Value Numbering

package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"

	"oooga.ooo/cs-1620/pkg/models"
	"oooga.ooo/cs-1620/pkg/utils"
)

func equalComputedValue(a, b models.Instruction) bool {
	// Blind pointer dereference is safe here because an instruction
	// wouldn't have a destination if it didn't have an operation. Getting
	// here in the code if we didn't have a destination should be
	// impossible.
	if *a.Op != *b.Op {
		return false
	}
	if (a.Value == nil) != (b.Value == nil) {
		return false
	}
	if a.Value != nil {
		if (a.Value.Float == nil) != (b.Value.Float == nil) {
			return false
		}
		if a.Value.Float != nil {
			if *a.Value.Float != *b.Value.Float {
				return false
			}
		}
		if (a.Value.Bool == nil) != (b.Value.Bool == nil) {
			return false
		}
		if a.Value.Bool != nil {
			if *a.Value.Bool != *b.Value.Bool {
				return false
			}
		}
	}
	if len(a.Args) != len(b.Args) {
		return false
	}
	for i := range a.Args {
		if a.Args[i] != b.Args[i] {
			return false
		}
	}
	return true
}

func equivalentComputationIndex(table []lvnTableEntry, instruction models.Instruction, env map[string]int, prop bool) (int, bool) {
	if prop && *instruction.Op == "id" {
		i, err := strconv.Atoi(instruction.Args[0])
		if err != nil {
			panic(err)
		}
		entry := table[i]
		if entry.inst != nil {
			return i, true
		}
	}
	for i, entry := range table {
		if entry.inst != nil {
			if equalComputedValue(instruction, *entry.inst) {
				return i, true
			}
		}
	}
	return -1, false
}

type lvnTableEntry struct {
	inst *models.Instruction
	cv   string // canonical value
}

var id = "id"

// lvn modifies instructions in place
func lvn(block []models.Instruction, prop bool) {
	// This stores a mapping between a computation that has taken place and
	// a variable with which it can be referred. Unlike the environment this
	// is a static map it doesn't update, new things are simply added. So a
	// variable in the table must always be able to refer to the computation
	// regardless of what the environment currently contains.
	var table []lvnTableEntry
	// This is our environment. It stores mappings between live variables
	// currently in our program and computations that have taken place. As
	// our program overwrites things variables in our environment will
	// change, while the computations don't.
	varToTableIdx := make(map[string]int)

	for blockIdx, inst := range block {
		if inst.Dest != nil {
			var argTableIdxs []int
			var mangledArgs []string
			for _, arg := range inst.Args {
				// If we don't have this in our environment it
				// came from a global context. Simple thing to
				// do here is just drop it in the environment
				// and table with no computation. It has a
				// computation we just don't know what it is.
				if _, ok := varToTableIdx[arg]; !ok {
					tableEntry := lvnTableEntry{
						cv: arg,
					}
					table = append(table, tableEntry)
					varToTableIdx[arg] = len(table) - 1
				}
				argTableIdxs = append(argTableIdxs, varToTableIdx[arg])
				mangledArgs = append(mangledArgs, strconv.Itoa(varToTableIdx[arg]))
				// For commutative operations consider sorted to
				// be canonical
				if *inst.Op == "add" || *inst.Op == "mul" {
					sort.Strings(mangledArgs)
				}
			}
			// We are reusing models.Instruction for our table value
			// even though it could probably be better expressed in
			// it's own type. Hence the mangled args.
			tableInst := inst
			tableInst.Args = mangledArgs
			if tableIdx, ok := equivalentComputationIndex(table, tableInst, varToTableIdx, prop); ok {
				foundInst := table[tableIdx]
				// equivalentComputationIndex will never return
				// table entry without operation blind
				// dereference is safe here
				op := *foundInst.inst.Op
				if prop && op == "id" || op == "const" {
					temp := *foundInst.inst
					temp.Dest = inst.Dest
					inst = temp
					var outArgs []string
					for _, arg := range inst.Args {
						ii, err := strconv.Atoi(arg)
						if err != nil {
							panic(err)
						}
						outArgs = append(outArgs, table[ii].cv)
					}
					inst.Args = outArgs
				} else {
					inst = models.Instruction{
						Args: []string{foundInst.cv},
						Dest: inst.Dest,
						Op:   &id,
						Type: inst.Type,
					}
				}
				varToTableIdx[*inst.Dest] = tableIdx
			} else {
				// Store original variable in environment so
				// proceeding uses can still look up variable by
				// original name not made up unique name (if we
				// end up making on). Use len(table) because the
				// entry is about to be added.
				varToTableIdx[*inst.Dest] = len(table)

				// Determine if variable is reused after this
				// point, if so give it unique name. This is
				// because the *value* may be re-used after the
				// variable is clobbered in the environment. In
				// that circumstance, we need a name to get at
				// this previously computed value. In the
				// environment the original name will point at
				// this entry in the table up until it gets
				// clobbered. Our pass will therefore re-write
				// all uses to our unique name in the final
				// output.
				for _, afterInst := range block[blockIdx+1:] {
					if afterInst.Dest != nil && *afterInst.Dest == *inst.Dest {
						uniqueDest := fmt.Sprintf("lvn.%d", blockIdx)
						inst.Dest = &uniqueDest
						break
					}
				}
				tableEntry := lvnTableEntry{
					inst: &tableInst,
					cv:   *inst.Dest, // we want to use the new name here
				}
				table = append(table, tableEntry)

				// Rewrite args for this function to use the
				// canonical variables
				for argIdx, tableIdx := range argTableIdxs {
					inst.Args[argIdx] = table[tableIdx].cv
				}
			}
		} else {
			for argIdx, arg := range inst.Args {
				tableEntry := table[varToTableIdx[arg]]
				if prop && *tableEntry.inst.Op == "id" {
					ii, err := strconv.Atoi(tableEntry.inst.Args[0])
					if err != nil {
						panic(err)
					}
					tableEntry = table[ii]
				}
				inst.Args[argIdx] = tableEntry.cv
			}
		}
		block[blockIdx] = inst
	}
}

func main() {
	prog := utils.ReadProgram()

	prop := flag.Bool("p", false, "")
	flag.Parse()
	println(*prop)

	for i, function := range prog.Functions {
		namesInOrder, nameToBlock := utils.BasicBlocks(function.Instrs)
		for _, blockName := range namesInOrder {
			// This will modify the block in place
			lvn(nameToBlock[blockName], *prop)
		}

		prog.Functions[i].Instrs = utils.FlattenBlocks(namesInOrder, nameToBlock)
	}

	utils.PrintProgram(prog)
}
