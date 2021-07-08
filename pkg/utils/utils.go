package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"oooga.ooo/cs-1620/pkg/models"
)

var terminators = [...]string{"jmp", "br", "ret"}

// BasicBlocks breaks program down into basic blocks which is just a term for a
// string of instructions with no control flow, just things that need to happen
// one after another.
func BasicBlocks(body []models.Instruction) (namesInOrder []string, nameToBlock map[string][]models.Instruction) {
	nameToBlock = make(map[string][]models.Instruction)

	var block []models.Instruction

	blockCounter := 1
	var blockName *string

	addBlock := func() {
		if len(block) != 0 {
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

type Digraph map[string][]string

// MakeCFG computes the control flow graph
func MakeCFG(namesInOrder []string, nameToBlock map[string][]models.Instruction) Digraph {
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

func Predecessors(cfg Digraph, name string) []string {
	var predecessors []string
	for n, to := range cfg {
		if contains(to, name) {
			predecessors = append(predecessors, n)
		}
	}
	return predecessors
}

func Successors(cfg Digraph, name string) []string {
	return cfg[name]
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

type Set map[string]struct{}

func (s Set) String() string {
	if len(s) == 0 {
		return "∅"
	}
	var items []string
	for k := range s {
		items = append(items, k)
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func (s Set) Add(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s Set) Remove(item string) {
	delete(s, item)
}

func Sub(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.Add(item)
	}
	for item := range lhs {
		out.Remove(item)
	}
	return out
}

func Union(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.Add(item)
	}
	for item := range lhs {
		out.Add(item)
	}
	return out
}

func Intersect(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		if lhs.Contains(item) {
			out.Add(item)
		}
	}
	return out
}
