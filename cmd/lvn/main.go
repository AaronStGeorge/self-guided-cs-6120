package main

import (
	"encoding/json"
	"fmt"
	"log"

	"oooga.ooo/cs-1620/pkg/utils"
)

func main() {
	prog := utils.ReadProgram()
	out, err := json.Marshal(&prog)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
}
