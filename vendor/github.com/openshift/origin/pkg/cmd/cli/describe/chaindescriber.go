package describe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/golang/glog"
	"github.com/gonum/graph"
	"github.com/gonum/graph/encoding/dot"
	"github.com/gonum/graph/path"
	kapi "k8s.io/kubernetes/pkg/api"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	"k8s.io/kubernetes/pkg/util/sets"

	osgraph "github.com/openshift/origin/pkg/api/graph"
	buildedges "github.com/openshift/origin/pkg/build/graph"
	buildanalysis "github.com/openshift/origin/pkg/build/graph/analysis"
	buildgraph "github.com/openshift/origin/pkg/build/graph/nodes"
	"github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/api"
	imagegraph "github.com/openshift/origin/pkg/image/graph/nodes"
	dotutil "github.com/openshift/origin/pkg/util/dot"
	"github.com/openshift/origin/pkg/util/parallel"
)

// NotFoundErr is returned when the imageStreamTag (ist) of interest cannot
// be found in the graph. This doesn't mean though that the IST does not
// exist. A user may have an image stream without a build configuration
// pointing at it. In that case, the IST of interest simply doesn't have
// other dependant ists
type NotFoundErr string

func (e NotFoundErr) Error() string {
	return fmt.Sprintf("couldn't find image stream tag: %q", string(e))
}

// ChainDescriber generates extended information about a chain of
// dependencies of an image stream
type ChainDescriber struct {
	c            client.BuildConfigsNamespacer
	namespaces   sets.String
	outputFormat string
	namer        osgraph.Namer
}

// NewChainDescriber returns a new ChainDescriber
func NewChainDescriber(c client.BuildConfigsNamespacer, namespaces sets.String, out string) *ChainDescriber {
	return &ChainDescriber{c: c, namespaces: namespaces, outputFormat: out, namer: namespacedFormatter{hideNamespace: true}}
}

// MakeGraph will create the graph of all build configurations and the image streams
// they point to via image change triggers in the provided namespace(s)
func (d *ChainDescriber) MakeGraph() (osgraph.Graph, error) {
	g := osgraph.New()

	loaders := []GraphLoader{}
	for namespace := range d.namespaces {
		glog.V(4).Infof("Loading build configurations from %q", namespace)
		loaders = append(loaders, &bcLoader{namespace: namespace, lister: d.c})
	}
	loadingFuncs := []func() error{}
	for _, loader := range loaders {
		loadingFuncs = append(loadingFuncs, loader.Load)
	}

	if errs := parallel.Run(loadingFuncs...); len(errs) > 0 {
		return g, utilerrors.NewAggregate(errs)
	}

	for _, loader := range loaders {
		loader.AddToGraph(g)
	}

	buildedges.AddAllInputOutputEdges(g)

	return g, nil
}

// Describe returns the output of the graph starting from the provided
// image stream tag (name:tag) in namespace. Namespace is needed here
// because image stream tags with the same name can be found across
// different namespaces.
func (d *ChainDescriber) Describe(ist *imageapi.ImageStreamTag, includeInputImages, reverse bool) (string, error) {
	g, err := d.MakeGraph()
	if err != nil {
		return "", err
	}

	// Retrieve the imageStreamTag node of interest
	istNode := g.Find(imagegraph.ImageStreamTagNodeName(ist))
	if istNode == nil {
		return "", NotFoundErr(fmt.Sprintf("%q", ist.Name))
	}

	markers := buildanalysis.FindCircularBuilds(g, d.namer)
	if len(markers) > 0 {
		for _, marker := range markers {
			if strings.Contains(marker.Message, ist.Name) {
				return marker.Message, nil
			}
		}
	}

	buildInputEdgeKinds := []string{buildedges.BuildTriggerImageEdgeKind}
	if includeInputImages {
		buildInputEdgeKinds = append(buildInputEdgeKinds, buildedges.BuildInputImageEdgeKind)
	}

	// Partition down to the subgraph containing the imagestreamtag of interest
	var partitioned osgraph.Graph
	if reverse {
		partitioned = partitionReverse(g, istNode, buildInputEdgeKinds)
	} else {
		partitioned = partition(g, istNode, buildInputEdgeKinds)
	}

	switch strings.ToLower(d.outputFormat) {
	case "dot":
		data, err := dot.Marshal(partitioned, dotutil.Quote(ist.Name), "", "  ", false)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "":
		return d.humanReadableOutput(partitioned, d.namer, istNode, reverse), nil
	}

	return "", fmt.Errorf("unknown specified format %q", d.outputFormat)
}

// partition the graph down to a subgraph starting from the given root
func partition(g osgraph.Graph, root graph.Node, buildInputEdgeKinds []string) osgraph.Graph {
	// Filter out all but BuildConfig and ImageStreamTag nodes
	nodeFn := osgraph.NodesOfKind(buildgraph.BuildConfigNodeKind, imagegraph.ImageStreamTagNodeKind)
	// Filter out all but BuildInputImage and BuildOutput edges
	edgeKinds := []string{}
	edgeKinds = append(edgeKinds, buildInputEdgeKinds...)
	edgeKinds = append(edgeKinds, buildedges.BuildOutputEdgeKind)
	edgeFn := osgraph.EdgesOfKind(edgeKinds...)
	sub := g.Subgraph(nodeFn, edgeFn)

	// Filter out inbound edges to the IST of interest
	edgeFn = osgraph.RemoveInboundEdges([]graph.Node{root})
	sub = sub.Subgraph(nodeFn, edgeFn)

	// Check all paths leading from the root node, collect any
	// node found in them, and create the desired subgraph
	desired := []graph.Node{root}
	paths := path.DijkstraAllPaths(sub)
	for _, node := range sub.Nodes() {
		if node == root {
			continue
		}
		path, _, _ := paths.Between(root, node)
		if len(path) != 0 {
			desired = append(desired, node)
		}
	}
	return sub.SubgraphWithNodes(desired, osgraph.ExistingDirectEdge)
}

// partitionReverse the graph down to a subgraph starting from the given root
func partitionReverse(g osgraph.Graph, root graph.Node, buildInputEdgeKinds []string) osgraph.Graph {
	// Filter out all but BuildConfig and ImageStreamTag nodes
	nodeFn := osgraph.NodesOfKind(buildgraph.BuildConfigNodeKind, imagegraph.ImageStreamTagNodeKind)
	// Filter out all but BuildInputImage and BuildOutput edges
	edgeKinds := []string{}
	edgeKinds = append(edgeKinds, buildInputEdgeKinds...)
	edgeKinds = append(edgeKinds, buildedges.BuildOutputEdgeKind)
	edgeFn := osgraph.EdgesOfKind(edgeKinds...)
	sub := g.Subgraph(nodeFn, edgeFn)

	// Filter out inbound edges to the IST of interest
	edgeFn = osgraph.RemoveOutboundEdges([]graph.Node{root})
	sub = sub.Subgraph(nodeFn, edgeFn)

	// Check all paths leading from the root node, collect any
	// node found in them, and create the desired subgraph
	desired := []graph.Node{root}
	paths := path.DijkstraAllPaths(sub)
	for _, node := range sub.Nodes() {
		if node == root {
			continue
		}
		path, _, _ := paths.Between(node, root)
		if len(path) != 0 {
			desired = append(desired, node)
		}
	}
	return sub.SubgraphWithNodes(desired, osgraph.ExistingDirectEdge)
}

// humanReadableOutput traverses the provided graph using DFS and outputs it
// in a human-readable format. It starts from the provided root, assuming it
// is an imageStreamTag node and continues to the rest of the graph handling
// only imageStreamTag and buildConfig nodes.
func (d *ChainDescriber) humanReadableOutput(g osgraph.Graph, f osgraph.Namer, root graph.Node, reverse bool) string {
	if reverse {
		g = g.EdgeSubgraph(osgraph.ReverseExistingDirectEdge)
	}

	var singleNamespace bool
	if len(d.namespaces) == 1 && !d.namespaces.Has(kapi.NamespaceAll) {
		singleNamespace = true
	}
	depth := map[graph.Node]int{
		root: 0,
	}
	out := ""

	dfs := &DepthFirst{
		Visit: func(u, v graph.Node) {
			depth[v] = depth[u] + 1
		},
	}

	until := func(node graph.Node) bool {
		var info string

		switch t := node.(type) {
		case *imagegraph.ImageStreamTagNode:
			info = outputHelper(f.ResourceName(t), t.Namespace, singleNamespace)
		case *buildgraph.BuildConfigNode:
			info = outputHelper(f.ResourceName(t), t.BuildConfig.Namespace, singleNamespace)
		default:
			panic("this graph contains node kinds other than imageStreamTags and buildConfigs")
		}

		if depth[node] != 0 {
			out += "\n"
		}
		out += fmt.Sprintf("%s", strings.Repeat("\t", depth[node]))
		out += fmt.Sprintf("%s", info)

		return false
	}

	dfs.Walk(g, root, until)

	return out
}

// outputHelper returns resource/name in a single namespace, <namespace resource/name>
// in multiple namespaces
func outputHelper(info, namespace string, singleNamespace bool) string {
	if singleNamespace {
		return info
	}
	return fmt.Sprintf("<%s %s>", namespace, info)
}

// DepthFirst implements stateful depth-first graph traversal.
// Modifies behavior of visitor.DepthFirst to allow nodes to be visited multiple
// times as long as they're not in the current stack
type DepthFirst struct {
	EdgeFilter func(graph.Edge) bool
	Visit      func(u, v graph.Node)
	stack      NodeStack
}

// Walk performs a depth-first traversal of the graph g starting from the given node
func (d *DepthFirst) Walk(g graph.Graph, from graph.Node, until func(graph.Node) bool) graph.Node {
	return d.visit(g, from, until)
}

func (d *DepthFirst) visit(g graph.Graph, t graph.Node, until func(graph.Node) bool) graph.Node {
	if until != nil && until(t) {
		return t
	}
	d.stack.Push(t)
	children := osgraph.ByID(g.From(t))
	sort.Sort(children)
	for _, n := range children {
		if d.EdgeFilter != nil && !d.EdgeFilter(g.Edge(t, n)) {
			continue
		}
		if d.visited(n.ID()) {
			continue
		}
		if d.Visit != nil {
			d.Visit(t, n)
		}
		result := d.visit(g, n, until)
		if result != nil {
			return result
		}
	}
	d.stack.Pop()
	return nil
}

func (d *DepthFirst) visited(id int) bool {
	for _, n := range d.stack {
		if n.ID() == id {
			return true
		}
	}
	return false
}

// NodeStack implements a LIFO stack of graph.Node.
// NodeStack is internal only in go 1.5.
type NodeStack []graph.Node

// Len returns the number of graph.Nodes on the stack.
func (s *NodeStack) Len() int { return len(*s) }

// Pop returns the last graph.Node on the stack and removes it
// from the stack.
func (s *NodeStack) Pop() graph.Node {
	v := *s
	v, n := v[:len(v)-1], v[len(v)-1]
	*s = v
	return n
}

// Push adds the node n to the stack at the last position.
func (s *NodeStack) Push(n graph.Node) { *s = append(*s, n) }
