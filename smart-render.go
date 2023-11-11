package animax

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Graph struct {
	Nodes map[string][]string
	Ordering []string
}

var VideoGraph []string = []string{
	"-filter_complex",
	"-aspect",
	"-filter:v|-filter:a",
	"-vf|-va",
}

func GetRenderGraph(graphRules []string) Graph {
	graph := Graph{
		Nodes: make(map[string][]string),
		Ordering: nil,
	}

	graph.loadRenderRules(&graphRules)
	// graph.ProduceOrdering()
	return graph
}

func (graph *Graph) loadRenderRules(graphRules *[]string) {
	for _, rule := range *graphRules {
		nodes :=strings.Split(rule, "|")
		for _, src := range nodes {
			for _, dest := range nodes {
				// fmt.Printf("SRC: %s | DEST: %s\n", src, dest)
				graph.AddEdge(src, dest)
			}
		}
	}
}

func remove[T any](slice *[]T, index int) T {
	item := (*slice)[index]
	*slice = append((*slice)[:index], (*slice)[index+1:]...) 
    return item
}

func processFilterComplex(args *Args) []string {
	tag := ""
	output := []string{"-filter_complex"}
	filter := ""

	for index, val := range (*args)["-filter_complex"] {
		if index == 0 && strings.HasPrefix("trim=", val.Value)  {
			tag = uuid.New().String()[0:4]
			filter += fmt.Sprintf(`[0]%s[%s];`, val.Value, tag)
			continue
		}
		if index == 0  {
			tag = uuid.New().String()[0:4]
			filter += fmt.Sprintf(`[0]%s[%s];`, val.Value, tag)
			continue
		}
		filter += fmt.Sprintf(`[%s]`, tag)
		tag = uuid.New().String()[0:4]
		filter += fmt.Sprintf(`%s[%s];`, val.Value, tag)
	}
	output = append(output, filter[0:len(filter) - 1])
																										//videoEncoding
	output = append(output, []string{"-map", "[" + tag + "]", "-map", "0:a", "-c:v", "libx264"}...)
	(*args)["-filter_complex"] = []subArg{}

	return output
}

// topological sort
func (g *Graph) ProduceOrdering(args Args) [][]string {
	visited := make(map[string]bool)
	renderStages := [][]string{}
	for node, _  := range g.Nodes {
		stage := []string{}

		all, ok := args[node]
		if ok && !visited[node] {
			visited[node] = true
			if node == "-filter_complex" {
				renderStages = append(renderStages, processFilterComplex(&args))
				continue
			}

			stage = append(stage, node)
			subArgValue := remove(&all, 0)
			stage = append(stage, subArgValue.Value)

			renderStages = append(renderStages, stage)
		}
	}

	fmt.Printf("%+v", renderStages)
	return renderStages
}

func (g *Graph) GetRenderStages(args Args) {

}

func (g *Graph) nodeExists(node string) bool {
	_, ok := g.Nodes[node]
	return ok
}

func (g *Graph) edgeNodeExists(srcNode string, destNode string) bool {
	for _, node := range g.Nodes[srcNode] {
		if node == destNode {return true}
	}

	return false
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
		if !g.edgeNodeExists(node1, node2) {
			g.Nodes[node1] = append(g.Nodes[node1], node2)
		}
	}

	if _, ok := g.Nodes[node2]; ok {
		if !g.edgeNodeExists(node2, node1) {
			g.Nodes[node2] = append(g.Nodes[node2], node1)
		}
	}
}

func (g Graph) PrintStages(args *Args) {

}

type set struct {
	s map[interface{}]bool
}
