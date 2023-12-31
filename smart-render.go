package animax

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type Graph struct {
	Nodes map[string][]string
	Ordering []string
}

var VideoGraph []string = []string{
	"-filter_complex",
	"-ss",
	"-aspect",
	"-filter:v|-filter:a",
	"-vf|-va",
}

var AudioGraph []string = []string{
	"-filter_complex",
	"-ss",
	"-af",
	"-vf|-va",
	"-filter:a|-filter:v",
}

func GetRenderGraph(graphRules []string) Graph {
	graph := Graph{
		Nodes: make(map[string][]string),
		Ordering: nil,
	}

	graph.loadRenderRules(&graphRules)
	return graph
}

func (graph *Graph) loadRenderRules(graphRules *[]string) {
	for _, rule := range *graphRules {
		nodes :=strings.Split(rule, "|")
		for _, src := range nodes {
			for _, dest := range nodes {
				graph.AddEdge(src, dest)
			}
		}
	}
}

func removeAtIndex[T any](slice *[]T, index int) {
	if index > len(*slice) - 1 {return}
	*slice = append((*slice)[:index], (*slice)[index+1:]...) 
}

func removeElement[T comparable](slice *[]T, element *T) []T {
    for i, v := range *slice {
        if reflect.DeepEqual(v, *element) {
            *slice = append((*slice)[:i], (*slice)[i+1:]...)
            return *slice
        }
    }

	return *slice
}
func processFilterComplex(args *Args, file *File) []string {
	tag := ""
	output := []string{"-filter_complex"}
	filter := ""
	set := newSet()
	toRemove := []*subArg{}
	for index, val := range (*args)["-filter_complex"] {
		// if index == 0 && strings.HasPrefix("trim=", val.Value)  {
		// 	tag = uuid.New().String()[0:4]
		// 	filter += fmt.Sprintf(`[0]%s[%s];`, val.Value, tag)
		// 	continue
		// }

		if index == 0  {
			tag = uuid.New().String()[0:4]
			filter += fmt.Sprintf(`[0]%s[%s];`, val.Value, tag)
			set.add(val.Key)
			toRemove = append(toRemove, &val)
			continue
		}

		if set.exists(val.Key) {continue}

		filter += fmt.Sprintf(`[%s]`, tag)
		tag = uuid.New().String()[0:4]
		filter += fmt.Sprintf(`%s[%s];`, val.Value, tag)
		set.add(val.Key)
		toRemove = append(toRemove, &val)
	}

	

	for i := 0; i < len(toRemove); i++ {
		tempVar := (*args)["-filter_complex"]
		// fmt.Printf("Before removal: %+v\n", (*args)["-filter_complex"])
		(*args)["-filter_complex"] = removeElement[subArg](&tempVar, toRemove[i] )
		// fmt.Println("IN FILTER")
		// fmt.Printf("After removal: %+v\n", (*args)["-filter_complex"])
		// time.Sleep(5)
	} 
	// fmt.Printf("After removal %+v", (*args)["-filter_complex"])
	fmt.Println("Filter" + filter)
	output = append(output, filter[0:len(filter) - 1])//filter[0:len(filter) - 1]
																										//videoEncoding
	switch (*file).GetType() {
		case video:
			output = append(output, []string{"-map", "[" + tag + "]", "-map", "0:a",}...)
			return output
		case audio:
			output = append(output, []string{"-map", "[" + tag + "]"}...)
			return output
	}

	// (*args)["-filter_complex"] = []subArg{}
	return output
}

func again(args *Args) bool {
	for _, v := range *args {
		if len(v) > 0 {return true}
	}


	return false
}

func prioritizeTrim(renderStages *[][]string) [][]string {
	trims := [][]string{}
	others := [][]string{}
	
	for _, val := range *renderStages {
		if isTrim(&val) {
			trims = append(trims, val)
			continue
		}	

		others = append(others, val)
	}

	return append(trims, others...)
}

// topological sort
func (g *Graph) ProduceOrdering(args Args, file File) [][]string {
    renderStages := [][]string{}
    for again(&args) {
		// fmt.Println("LOOP runing")
        visited := make(map[string]bool)

        renders := len(renderStages)
        for node, neighbors := range g.Nodes {
            stage := []string{}

            all, ok := args[node]
            if ok && !visited[node] {

                if node == "-filter_complex" {
                    renderStages = append(renderStages, processFilterComplex(&args, &file))
                    visited[node] = true
                    continue
                }

                if len(all) > 0 {
                    stage = append(stage, node)
                    subArgValue := all[0]
                    fmt.Printf("Before removal %+v \n", all)
                    temp := args[node]
                    removeAtIndex(&temp, 0)
					fmt.Printf("Temp after %+v\n", temp)
                    args[node] = temp
                    fmt.Printf("After removal %+v \n", args[node] )

                    stage = append(stage, subArgValue.Value)
                    visited[node] = true

                    for _, currentNeighbor := range neighbors {
                        a, ok := args[currentNeighbor]
                        if ok && !visited[currentNeighbor] {
                            visited[currentNeighbor] = true

                            if len(a) == 0 {
                                continue
                            }

                            stage = append(stage, currentNeighbor)
                            subArgValue := a[0]
                            temp := args[currentNeighbor]
                            removeAtIndex(&temp, 0)
                            args[currentNeighbor] = temp // Update the original map
                            stage = append(stage, subArgValue.Value)
                        }
                    }

                }

                renderStages = append(renderStages, stage)
            }
        }

        if renders == len(renderStages) {
            break
        }
    }

	renderStages = prioritizeTrim(&renderStages)

    return renderStages
}

// func (g *Graph) GetRenderStages(args Args) {

// }

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

type hashset struct {
	s map[interface{}]bool
}


func newSet() hashset {
	return hashset{
		s: make(map[interface{}]bool),
	}
}

func (set *hashset) add(key string) {
	set.s[key] = true
}

func (set *hashset) remove(key string) {
	delete(set.s, key)
}

func (set hashset) exists(key string) bool {
	if _, ok := set.s[key]; ok {
		return true
	}

	return false
}