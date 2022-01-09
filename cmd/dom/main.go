package main

import (
	"os"

	"aaronstgeorge.com/self-guided-cs-1620/pkg/dominators"

	"aaronstgeorge.com/self-guided-cs-1620/pkg/utils"
)

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		println("usage: dom command")
		os.Exit(1)
	}

	prog := utils.ReadProgram()

	//// the [0] is definitely not a reasonable thing to do in a production circumstance
	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.CFG(namesInOrder, nameToBlock)

	switch args[0] {
	case "dom":
		nameToDominators := dominators.Dominators(namesInOrder, nameToBlock, cfg)
		utils.OutputBlockNameToSet(namesInOrder, nameToDominators)
	case "tree":
		nameToDominators := dominators.Dominators(namesInOrder, nameToBlock, cfg)
		utils.OutputDot(namesInOrder, dominators.Tree(namesInOrder, cfg, nameToDominators))
	case "front":
		nameToDominators := dominators.Dominators(namesInOrder, nameToBlock, cfg)
		domTree := dominators.Tree(namesInOrder, cfg, nameToDominators)
		front := dominators.Front(namesInOrder, cfg, domTree)
		utils.OutputBlockNameToSet(namesInOrder, front)
	default:
		println("unknown command")
		os.Exit(1)
	}
}
