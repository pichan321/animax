package animax

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

type Graph struct {
	Nodes map[string][]string
}

const (
	VIDEO string = "video-rules.txt"
	AUDIO string = "audio-rules.txt"
)


func GetRenderGraph(pathToRenderRules string) Graph {
	graph := Graph{
		Nodes: make(map[string][]string),
	}

	graph.LoadRenderRules(pathToRenderRules)

	return graph
}

func (graph *Graph)LoadRenderRules(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("invalid rule file path")
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}

func (g Graph) nodeExists(node string) bool {
	_, ok := g.Nodes[node]
	return ok
}

func (g *Graph) AddNode(node string) {
	if g.nodeExists(node) {
		return
	}

	g.Nodes[node] = []string{}
}

func (g *Graph) AddEdge(node1 string, node2 string) {
	if node1 == node2 {
		g.AddNode(node1)
		return
	}

	g.AddNode(node1)
	g.AddNode(node2)

	if _, ok := g.Nodes[node1]; ok {
		g.Nodes[node1] = append(g.Nodes[node1], node2)
	}

	if _, ok := g.Nodes[node2]; ok {
		g.Nodes[node2] = append(g.Nodes[node2], node1)
	}
}

func (g Graph) PrintStages(args *Args) {

}