package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"oooga.ooo/cs-1620/pkg/models"
)

var terminators = [...]string{"jmp", "br", "ret"}

// BasicBlocks breaks program down into basic blocks which is just a term for a
// string of instructions with no control flow, just things that need to happen
// one after another.
func BasicBlocks(function models.Function) (namesInOrder []string, nameToBlock map[string][]models.Instruction) {
	nameToBlock = make(map[string][]models.Instruction)

	var block []models.Instruction

	blockCounter := 1
	var blockName *string

	addBlock := func() {
		// If there was no blockName from a label for the block give it one
		if blockName == nil {
			tmp := fmt.Sprintf("b%d", blockCounter)
			blockName = &tmp
			blockCounter++
		}

		// Add to output
		nameToBlock[*blockName] = block
		namesInOrder = append(namesInOrder, *blockName)

		// Reset state
		block = []models.Instruction{}
		blockName = nil
	}

	for _, instruction := range function.Instrs {
		if instruction.Op != nil {
			block = append(block, instruction)
			if contains(terminators[:], *instruction.Op) {
				addBlock()
			}
		} else { // we have a label
			if len(block) != 0 {
				addBlock()
			} else if len(namesInOrder) == 0 && len(function.Args) != 0 {
				// If the first thing in the function is a label
				// and the function has an argument create a
				// block. Jumping to the label is not the same
				// as jumping to the entry point of the function
				// given that the has an argument that comes
				// into existence in that block... Maybe, this
				// doesn't really make intuitive sense to me,
				// it's just a case that came up in the existing
				// tests. That explanation sort of makes sense
				// though.
				temp := "entry1"
				blockName = &temp
				addBlock()
			}
			// The next block will start with label we just found
			blockName = instruction.Label
			block = append(block, instruction)
		}
	}

	addBlock()
	return namesInOrder, nameToBlock
}

func contains(strs []string, str string) bool {
	for _, a := range strs {
		if a == str {
			return true
		}
	}
	return false
}

// CFG computes the control flow graph
func CFG(namesInOrder []string, nameToBlock map[string][]models.Instruction) Digraph {
	nameToJumpedTo := make(map[string][]string)
	for i, name := range namesInOrder {
		block := nameToBlock[name]
		var jumpedTo []string

		proceedingBlock := func() {
			if i != len(namesInOrder)-1 {
				jumpedTo = append(jumpedTo, namesInOrder[i+1])
			}
		}

		if len(block) == 0 {
			proceedingBlock()
		} else {
			// If the last instruction is a jmp or a br then the jumped to
			// blocks are whatever the labels are for that instruction.
			lastInst := block[len(block)-1]
			switch *lastInst.Op {
			case "jmp", "br":
				jumpedTo = lastInst.Labels
			case "ret":
				// Return instructions don't have following instructions.
			default:
				// If we're not at the last block and it's not a jmp instruction
				// then the next block in the control flow graph is just the
				// proceeding block.
				proceedingBlock()
			}
		}

		if len(jumpedTo) != 0 {
			nameToJumpedTo[name] = jumpedTo
		}
	}
	return nameToJumpedTo
}

// ReadProgram reads program from STDIN. All errors are fatal.
func ReadProgram() models.Program {
	reader := bufio.NewReader(os.Stdin)
	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	var prog models.Program
	err := json.Unmarshal([]byte(string(output)), &prog)
	if err != nil {
		log.Fatal(err)
	}

	return prog
}

func PrintProgram(prog models.Program) {
	out, err := json.Marshal(&prog)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func FlattenBlocks(namesInOrder []string, nameToBlock map[string][]models.Instruction) []models.Instruction {
	var out []models.Instruction
	for _, name := range namesInOrder {
		out = append(out, nameToBlock[name]...)
	}
	return out
}

func AddRet(namesInOrder []string, nameToBlock map[string][]models.Instruction) ([]string, map[string][]models.Instruction) {
	out := make(map[string][]models.Instruction)
	for _, name := range namesInOrder {
		out[name] = nameToBlock[name]
	}
	ret := "ret"
	inst := models.Instruction{Op: &ret}
	last := namesInOrder[len(namesInOrder)-1]
	out[last] = append(out[last], inst)
	return namesInOrder, out
}

func LabelNonEmptyBlocks(namesInOrder []string, nameToBlock map[string][]models.Instruction) ([]string, map[string][]models.Instruction) {
	//out := make(map[string][]models.Instruction)
	//for _, name := range namesInOrder {
	//	block := nameToBlock[name]
	//	if len(block) > 0 {
	//		if block[0].Label == nil {
	//			inst := models.Instruction{
	//				Label: &name,
	//			}
	//			block = append([]models.Instruction{inst}, block...)
	//		}
	//	}
	//	out[name] = block
	//}
	return namesInOrder, nameToBlock
}

func OutputBlockNameToSet(namesInOrder []string, nameToSet map[string]Set) {
	for _, name := range namesInOrder {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  %s\n", nameToSet[name])
	}
}
