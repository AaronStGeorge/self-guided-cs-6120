package main

import (
	"oooga.ooo/cs-1620/pkg/utils"
)

func main() {
	prog := utils.ReadProgram()

	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.MakeCFG(namesInOrder, nameToBlock)

	utils.OutputDot(namesInOrder, cfg)
}
