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

func MakeBlocks(body []models.Instruction) ([]string, map[string][]models.Instruction) {
	var namesInOrder []string
	nameToBlock := make(map[string][]models.Instruction)
	var block []models.Instruction
	b := 0
	var name *string
	addBlock := func() {
		if len(block) != 0 {
			// If there was no name from a label for the block give it one
			if name == nil {
				tmp := fmt.Sprintf("b%d", b)
				name = &tmp
				b++
			}
			nameToBlock[*name] = block
			namesInOrder = append(namesInOrder, *name)
			block = []models.Instruction{}
			name = nil
		}
	}

	for _, instruction := range body {
		if instruction.Op != nil {
			block = append(block, instruction)
			if contains(terminators[:], *instruction.Op) {
				addBlock()
			}
		} else { // we have a label
			addBlock()
			// The next block will start with label we just found
			name = instruction.Label
			block = append(block, instruction)
		}
	}

	addBlock()
	return namesInOrder, nameToBlock
}

func contains(instructions []string, instruction string) bool {
	for _, a := range instructions {
		if a == instruction {
			return true
		}
	}
	return false
}

// MakeCFG computes the control flow graph
func MakeCFG(namesInOrder []string, nameToBlock map[string][]models.Instruction) map[string][]string {
	nameToJumpedTo := make(map[string][]string)
	for i, name := range namesInOrder {
		block := nameToBlock[name]
		var jumpedTo []string

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
			if i != len(namesInOrder)-1 {
				jumpedTo = append(jumpedTo, namesInOrder[i+1])
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
