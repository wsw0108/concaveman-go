package concaveman

import (
	"math"
	"sort"

	"github.com/tidwall/tinyqueue"
	"github.com/wsw0108/concaveman/predicates"
	"github.com/wsw0108/concaveman/rbush"
)

type Options struct {
	Concavity       float64
	LengthThreshold float64
}

type node struct {
	p    Point
	prev *node
	next *node
	minX float64
	minY float64
	maxX float64
	maxY float64
}

// impl rbush.Item
func (p Point) Rect() (min, max [2]float64) {
	min = p
	max = p
	return
}

// impl rbush.Item
func (n node) Rect() (min, max [2]float64) {
	min = [2]float64{n.minX, n.minY}
	max = [2]float64{n.maxX, n.maxY}
	return
}

var (
	_ rbush.Item = Point{}
	_ rbush.Item = node{}
)

type qnode struct {
	node interface{}
	dist float64
}

// impl tinyqueue.Item
func (a qnode) Less(b tinyqueue.Item) bool {
	// compareDist(a, b)
	return a.dist < b.(*qnode).dist
}

func Concaveman(points []Point, opts ...Options) []Point {
	opt := Options{
		Concavity:       2,
		LengthThreshold: 0,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// a relative measure of concavity; higher value means simpler hull
	concavity := math.Max(0, opt.Concavity)
	// when a segment goes below this length threshold, it won't be drilled down further
	lengthThreshold := opt.LengthThreshold

	// start with a convex hull of the points
	hull := fastConvexHull(points)

	// index the points with an R-tree
	tree := rbush.New(16)
	items := make([]rbush.Item, 0, len(points))
	for _, p := range points {
		items = append(items, p)
	}
	tree.Load(items)

	// turn the convex hull into a linked list and populate the initial edge queue with the nodes
	queue := make([]*node, 0, len(hull))
	var last *node
	for _, p := range hull {
		tree.Remove(p)
		last = insertNode(p, last)
		queue = append(queue, last)
	}

	// index the segments with an R-tree (for intersection checks)
	// segTree := &rtree.RTreeG[*node]{}
	segTree := rbush.New(16)
	for _, n := range queue {
		updateBBox(n)
		// segTree.Insert([2]float64{n.minX, n.minY}, [2]float64{n.maxX, n.maxY}, n)
		segTree.Insert(n)
	}

	sqConcavity := concavity * concavity
	sqLenThreshold := lengthThreshold * lengthThreshold

	// process edges one by one
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		a := node.p
		b := node.next.p

		// skip the edge if it's already short enough
		sqLen := getSqDist(a, b)
		if sqLen < sqLenThreshold {
			continue
		}

		maxSqLen := sqLen / sqConcavity

		// find the best connection point for the current edge to flex inward to
		p, ok := findCandidate(tree, node.prev.p, a, b, node.next.next.p, maxSqLen, segTree)

		// if we found a connection and it satisfies our concavity measure
		if ok && math.Min(getSqDist(p, a), getSqDist(p, b)) <= maxSqLen {
			// connect the edge endpoints through this point and add 2 new edges to the queue
			queue = append(queue, node)
			queue = append(queue, insertNode(p, node))

			// update point and segment indexes
			tree.Remove(p)
			// segTree.Delete([2]float64{node.minX, node.minY}, [2]float64{node.maxX, node.maxY}, node)
			segTree.Remove(node)
			n1 := updateBBox(node)
			n2 := updateBBox(node.next)
			// segTree.Insert([2]float64{n1.minX, n1.minY}, [2]float64{n1.maxX, n1.maxY}, n1)
			// segTree.Insert([2]float64{n2.minX, n2.minY}, [2]float64{n2.maxX, n2.maxY}, n2)
			segTree.Insert(n1)
			segTree.Insert(n2)
		}
	}

	// convert the resulting hull linked list to an array of points
	node := last
	var concave []Point
	for {
		concave = append(concave, node.p)
		node = node.next
		if node == last {
			break
		}
	}

	concave = append(concave, node.p)

	return concave
}

// func findCandidate(tree *rbush.RBush, a, b, c, d Point, maxDist float64, segTree *rtree.RTreeG[*node]) (Point, bool) {
func findCandidate(tree *rbush.RBush, a, b, c, d Point, maxDist float64, segTree *rbush.RBush) (Point, bool) {
	queue := tinyqueue.New(nil)
	node := tree.Data

	// search through the point R-tree with a depth-first search using a priority queue
	// in the order of distance to the edge (b, c)
	for node != nil {
		for _, child := range node.Children {
			var dist float64
			if node.Leaf {
				dist = sqSegDist(child.(Point), b, c)
			} else {
				dist = sqSegBoxDist(b, c, child.(*rbush.TreeNode))
			}
			if dist > maxDist {
				// skip the node if it's farther than we ever need
				continue
			}
			queue.Push(&qnode{
				node: child,
				dist: dist,
			})
		}

		for queue.Len() > 0 {
			{
				item := queue.Peek()
				qn := item.(*qnode)
				if _, ok := qn.node.(Point); !ok {
					break
				}
			}
			item := queue.Pop()
			qn := item.(*qnode)
			p := qn.node.(Point)

			// skip all points that are as close to adjacent edges (a,b) and (c,d),
			// and points that would introduce self-intersections when connected
			d0 := sqSegDist(p, a, b)
			d1 := sqSegDist(p, c, d)
			if qn.dist < d0 && qn.dist < d1 &&
				noIntersections(b, p, segTree) &&
				noIntersections(c, p, segTree) {
				return p, true
			}
		}

		item := queue.Pop()
		if item != nil {
			qn := item.(*qnode)
			node = qn.node.(*rbush.TreeNode)
		} else {
			node = nil
		}
	}

	return Point{}, false
}

// square distance from a segment bounding box to the given one
func sqSegBoxDist(a, b Point, bbox *rbush.TreeNode) float64 {
	if inside(a, bbox) || inside(b, bbox) {
		return 0
	}
	d1 := sqSegSegDist(a[0], a[1], b[0], b[1], bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Min[1])
	if d1 == 0 {
		return 0
	}
	d2 := sqSegSegDist(a[0], a[1], b[0], b[1], bbox.Min[0], bbox.Min[1], bbox.Min[0], bbox.Max[1])
	if d2 == 0 {
		return 0
	}
	d3 := sqSegSegDist(a[0], a[1], b[0], b[1], bbox.Max[0], bbox.Min[1], bbox.Max[0], bbox.Max[1])
	if d3 == 0 {
		return 0
	}
	d4 := sqSegSegDist(a[0], a[1], b[0], b[1], bbox.Min[0], bbox.Max[1], bbox.Max[0], bbox.Max[1])
	if d4 == 0 {
		return 0
	}
	m1 := math.Min(d1, d2)
	m2 := math.Min(d3, d4)
	return math.Min(m1, m2)
}

func inside(a Point, bbox *rbush.TreeNode) bool {
	return a[0] >= bbox.Min[0] &&
		a[0] <= bbox.Max[0] &&
		a[1] >= bbox.Min[1] &&
		a[1] <= bbox.Max[1]
}

// check if the edge (a,b) doesn't intersect any other edges
// func noIntersections(a, b Point, segTree *rtree.RTreeG[*node]) bool {
func noIntersections(a, b Point, segTree *rbush.RBush) bool {
	minX := math.Min(a[0], b[0])
	minY := math.Min(a[1], b[1])
	maxX := math.Max(a[0], b[0])
	maxY := math.Max(a[1], b[1])

	var edges []*node

	// segTree.Search([2]float64{minX, minY}, [2]float64{maxX, maxY}, func(_, _ [2]float64, data *node) bool {
	// 	edges = append(edges, data)
	// 	return true
	// })

	s := node{
		minX: minX,
		minY: minY,
		maxX: maxX,
		maxY: maxY,
	}
	segTree.Search(s, func(item rbush.Item) bool {
		edge := item.(*node)
		edges = append(edges, edge)
		return true
	})

	for _, edge := range edges {
		if intersects(edge.p, edge.next.p, a, b) {
			return false
		}
	}
	return true
}

func cross(p1, p2, p3 Point) float64 {
	return predicates.Orient2D(p1[0], p1[1], p2[0], p2[1], p3[0], p3[1])
}

// check if the edges (p1,q1) and (p2,q2) intersect
func intersects(p1, q1, p2, q2 Point) bool {
	return (p1[0] != q2[0] || p1[1] != q2[1]) &&
		(q1[0] != p2[0] || q1[1] != p2[1]) &&
		(cross(p1, q1, p2) > 0) != (cross(p1, q1, q2) > 0) &&
		(cross(p2, q2, p1) > 0) != (cross(p2, q2, q1) > 0)
}

// update the bounding box of a node's edge
func updateBBox(node *node) *node {
	p1 := node.p
	p2 := node.next.p
	node.minX = math.Min(p1[0], p2[0])
	node.minY = math.Min(p1[1], p2[1])
	node.maxX = math.Max(p1[0], p2[0])
	node.maxY = math.Max(p1[1], p2[1])
	return node
}

// speed up convex hull by filtering out points inside quadrilateral formed by 4 extreme points
func fastConvexHull(points []Point) []Point {
	left := points[0]
	top := points[0]
	right := points[0]
	bottom := points[0]

	// find the leftmost, rightmost, topmost and bottommost points
	for _, p := range points {
		if p[0] < left[0] {
			left = p
		}
		if p[0] > right[0] {
			right = p
		}
		if p[1] < top[1] {
			top = p
		}
		if p[1] > bottom[1] {
			bottom = p
		}
	}

	// filter out points that are inside the resulting quadrilateral
	cull := []Point{left, top, right, bottom}
	filtered := []Point{left, top, right, bottom}
	for _, p := range points {
		if !PointInPolygon(p, cull) {
			filtered = append(filtered, p)
		}
	}

	// get convex hull around the filtered points
	return convexHull(filtered)
}

// create a new node in a doubly linked list
func insertNode(p Point, prev *node) *node {
	node := &node{
		p: p,
	}

	if prev == nil {
		node.prev = node
		node.next = node
	} else {
		node.next = prev.next
		node.prev = prev
		prev.next.prev = node
		prev.next = node
	}

	return node
}

// square distance between 2 points
func getSqDist(p1, p2 Point) float64 {
	dx := p1[0] - p2[0]
	dy := p1[1] - p2[1]

	return dx*dx + dy*dy
}

// segment to segment distance, ported from http://geomalgorithms.com/a07-_distance.html by Dan Sunday
func sqSegSegDist(x0, y0, x1, y1, x2, y2, x3, y3 float64) float64 {
	ux := x1 - x0
	uy := y1 - y0
	vx := x3 - x2
	vy := y3 - y2
	wx := x0 - x2
	wy := y0 - y2
	a := ux*ux + uy*uy
	b := ux*vx + uy*vy
	c := vx*vx + vy*vy
	d := ux*wx + uy*wy
	e := vx*wx + vy*wy
	D := a*c - b*b

	var sc, sN, tc, tN float64
	sD := D
	tD := D

	if D == 0 {
		sN = 0
		sD = 1
		tN = e
		tD = c
	} else {
		sN = b*e - c*d
		tN = a*e - b*d
		if sN < 0 {
			sN = 0
			tN = e
			tD = c
		} else if sN > sD {
			sN = sD
			tN = e + b
			tD = c
		}
	}

	if tN < 0.0 {
		tN = 0.0
		if -d < 0.0 {
			sN = 0.0
		} else if -d > a {
			sN = sD
		} else {
			sN = -d
			sD = a
		}
	} else if tN > tD {
		tN = tD
		if (-d + b) < 0.0 {
			sN = 0
		} else if -d+b > a {
			sN = sD
		} else {
			sN = -d + b
			sD = a
		}
	}

	if sN == 0 {
		sc = 0
	} else {
		sc = sN / sD
	}
	if tN == 0 {
		tc = 0
	} else {
		tc = tN / tD
	}

	cx := (1-sc)*x0 + sc*x1
	cy := (1-sc)*y0 + sc*y1
	cx2 := (1-tc)*x2 + tc*x3
	cy2 := (1-tc)*y2 + tc*y3
	dx := cx2 - cx
	dy := cy2 - cy

	return dx*dx + dy*dy
}

// square distance from a point to a segment
func sqSegDist(p, p1, p2 Point) float64 {
	x := p1[0]
	y := p1[1]
	dx := p2[0] - x
	dy := p2[1] - y

	if dx != 0 || dy != 0 {

		t := ((p[0]-x)*dx + (p[1]-y)*dy) / (dx*dx + dy*dy)

		if t > 1 {
			x = p2[0]
			y = p2[1]

		} else if t > 0 {
			x += dx * t
			y += dy * t
		}
	}

	dx = p[0] - x
	dy = p[1] - y

	return dx*dx + dy*dy
}

type compareByX []Point

func (ps compareByX) Len() int {
	return len(ps)
}

func f64Less(a, b float64) bool {
	return a < b || (math.IsNaN(a) && !math.IsNaN(b))
}

func (ps compareByX) Less(i, j int) bool {
	if ps[i][0] == ps[j][0] {
		return f64Less(ps[i][1], ps[j][1])
	}
	return f64Less(ps[i][0], ps[j][0])
}

func (ps compareByX) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func convexHull(points []Point) []Point {
	sort.Sort(compareByX(points))

	var lower []Point
	for i := range points {
		p := points[i]
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}

	var upper []Point
	for i := range points {
		p := points[len(points)-i-1]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}

	result := make([]Point, 0, len(lower)-1+len(upper)-1)
	for i := range lower {
		if i == len(lower)-1 {
			break
		}
		result = append(result, lower[i])
	}
	for i := range upper {
		if i == len(upper)-1 {
			break
		}
		result = append(result, upper[i])
	}

	return result
}
