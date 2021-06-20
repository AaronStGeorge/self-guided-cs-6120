// Local Value Numbering

package main

import (
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

func equivilantComputationIndex(table []lvnTableEntry, instruction models.Instruction) (int, bool) {
	for i, entry := range table {
		if equalComputedValue(instruction, entry.inst) {
			return i, true
		}
	}
	return -1, false
}

type lvnTableEntry struct {
	inst models.Instruction
	cv   string // canonical value
}

var id = "id"

// lvn modifies instructions in place
func lvn(block []models.Instruction) {

	var table []lvnTableEntry
	varToIdx := make(map[string]int)

	for blockIdx, inst := range block {
		if inst.Dest != nil {
			var argTableIdxs []int
			var mangledArgs []string
			for _, arg := range inst.Args {
				argTableIdxs = append(argTableIdxs, varToIdx[arg])
				mangledArgs = append(mangledArgs, strconv.Itoa(varToIdx[arg]))
			}
			// We are reusing models.Instruction for our table value
			// even though it could probably be better expressed in
			// it's own type. Hence the mangled args.
			tableInst := inst
			tableInst.Args = mangledArgs
			if tableIdx, ok := equivilantComputationIndex(table, tableInst); ok {
				inst = models.Instruction{
					Args: []string{table[tableIdx].cv},
					Dest: inst.Dest,
					Op:   &id,
					Type: inst.Type,
				}
				varToIdx[*inst.Dest] = tableIdx
			} else {
				tableEntry := lvnTableEntry{
					inst: tableInst,
					cv:   *inst.Dest,
				}
				table = append(table, tableEntry)
				varToIdx[*inst.Dest] = len(table) - 1

				// Rewrite args for this function to point at
				// canonical variables
				for argIdx, tableIdx := range argTableIdxs {
					inst.Args[argIdx] = table[tableIdx].cv
				}
			}
			block[blockIdx] = inst
		}
	}
}

func main() {
	prog := utils.ReadProgram()

	for i, function := range prog.Functions {
		namesInOrder, nameToBlock := utils.BasicBlocks(function.Instrs)
		for _, blockName := range namesInOrder {
			// This will modify
			lvn(nameToBlock[blockName])
		}

		prog.Functions[i].Instrs = utils.FlattenBlocks(namesInOrder, nameToBlock)
	}

	utils.PrintProgram(prog)
}
