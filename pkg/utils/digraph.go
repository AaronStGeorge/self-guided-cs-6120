package utils

import (
	"fmt"
	"sort"
)

type Digraph map[string][]string

func Predecessors(cfg Digraph, name string) []string {
	var predecessors []string
	for n, to := range cfg {
		if contains(to, name) {
			predecessors = append(predecessors, n)
		}
	}
	sort.Strings(predecessors)
	return predecessors
}

func Successors(cfg Digraph, name string) []string {
	return cfg[name]
}

type direction int

const (
	Up direction = iota
	Down
)

func bfs(cfg Digraph, start string, dir direction, walk func(name string) bool) bool {
	var nodes []string
	switch dir {
	case Up:
		nodes = Predecessors(cfg, start)
	case Down:
		nodes = Successors(cfg, start)
	}
	for _, name := range nodes {
		if walk(name) {
			return true
		}
	}
	for _, name := range nodes {
		if bfs(cfg, name, dir, walk) {
			return true
		}
	}
	return false
}

func WalkUp(cfg Digraph, start string, walk func(name string) bool) {
	bfs(cfg, start, Up, walk)
}

func OutputDot(namesInOrder []string, cfg Digraph) {
	fmt.Println("digraph G {")
	for _, name := range namesInOrder {
		sort.Strings(cfg[name])
		for _, jumped := range cfg[name] {
			fmt.Printf("  \"%s\" -> \"%s\";\n", name, jumped)
		}
	}
	fmt.Println("}")
}
