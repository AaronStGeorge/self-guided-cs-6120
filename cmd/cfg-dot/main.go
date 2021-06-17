package main

import (
	"fmt"

	"oooga.ooo/cs-1620/pkg/utils"
)

func outputDot(namesInOrder []string, cfg map[string][]string) {
	fmt.Println("digraph G {")
	for _, name := range namesInOrder {
		for _, jmpedTo := range cfg[name] {
			fmt.Printf("  %s -> %s;\n", name, jmpedTo)
		}
	}
	fmt.Println("}")
}

func main() {
	prog := utils.ProgramFromStdin()

	namesInOrder, nameToBlock := utils.MakeBlocks(prog.Functions[0].Instrs)
	cfg := utils.MakeCFG(namesInOrder, nameToBlock)

	outputDot(namesInOrder, cfg)
}
