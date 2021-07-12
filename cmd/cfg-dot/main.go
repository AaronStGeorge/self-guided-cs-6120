package main

import (
	"oooga.ooo/cs-1620/pkg/utils"
)

func main() {
	prog := utils.ReadProgram()

	namesInOrder, nameToBlock := utils.BasicBlocks(prog.Functions[0])
	cfg := utils.CFG(namesInOrder, nameToBlock)

	utils.OutputDot(namesInOrder, cfg)
}
