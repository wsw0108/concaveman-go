package rbush

import (
	"math"
	"sort"
)

var (
	mathInfNeg = math.Inf(-1)
	mathInfPos = math.Inf(+1)
)

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

type TreeNode struct {
	Min, Max [2]float64
	Children []interface{}
	Leaf     bool
	height   int
}

type Item interface {
	Rect() (min, max [2]float64)
}

type RBush struct {
	maxEntries int
	minEntries int
	Data       *TreeNode
	reusePath  []*TreeNode
}

func New(maxEntries int) *RBush {
	tr := &RBush{}
	tr.maxEntries = int(mathMax(4, float64(maxEntries)))
	tr.minEntries = int(mathMax(2, math.Ceil(float64(tr.maxEntries)*0.4)))
	tr.Clear()
	return tr
}

func fillBBox(item Item, bbox *TreeNode) {
	bbox.Min, bbox.Max = item.Rect()
}

func (tr *RBush) Search(bbox Item, iter func(item Item) bool) bool {
	if bbox == nil {
		panic("bbox is nil")
	}
	min, max := bbox.Rect()
	return tr.searchBBox(min, max, iter)
}

func (tr *RBush) searchBBox(min, max [2]float64, iter func(item Item) bool) bool {
	bbox := TreeNode{Min: min, Max: max}
	if !tr.Data.intersects(&bbox) {
		return true
	}
	return search(tr.Data, &bbox, iter)
}

func search(node, bbox *TreeNode, iter func(item Item) bool) bool {
	if node.Leaf {
		for i := 0; i < len(node.Children); i++ {
			item := node.Children[i].(Item)
			var childBBox TreeNode
			fillBBox(item, &childBBox)
			if bbox.intersects(&childBBox) {
				if !iter(item) {
					return false
				}
			}
		}
	} else {
		for i := 0; i < len(node.Children); i++ {
			childBBox := node.Children[i].(*TreeNode)
			if bbox.intersects(childBBox) {
				if !search(childBBox, bbox, iter) {
					return false
				}
			}
		}
	}
	return true
}

func (tr *RBush) Load(data []Item) {
	if len(data) < tr.minEntries {
		for _, item := range data {
			tr.Insert(item)
		}
		return
	}

	// data.slice()?
	node := tr.build(data, 0, len(data)-1, 0)

	if len(tr.Data.Children) == 0 {
		tr.Data = node
	} else if tr.Data.height == node.height {
		tr.splitRoot(tr.Data, node)
	} else {
		if tr.Data.height < node.height {
			tr.Data, node = node, tr.Data
		}
		tr.insertNode(node, tr.Data.height-node.height-1)
	}
}

func (tr *RBush) Insert(item Item) {
	if item == nil {
		panic("item is nil")
	}
	tr.insertItem(item)
}

func (tr *RBush) Clear() {
	tr.Data = createNode(nil)
}

func (tr *RBush) Remove(item Item) {
	if item == nil {
		panic("item is nil")
	}

	node := tr.Data

	var bbox TreeNode
	fillBBox(item, &bbox)

	path := tr.reusePath[:0]

	var indexes []int

	var i int
	var parent *TreeNode
	var goingUp bool

	for node != nil || len(path) > 0 {
		if node == nil {
			node = path[len(path)-1]
			path = path[:len(path)-1]
			if len(path) == 0 {
				parent = nil
			} else {
				parent = path[len(path)-1]
			}
			i = indexes[len(indexes)-1]
			indexes = indexes[:len(indexes)-1]
			goingUp = true
		}

		if node.Leaf {
			index := findItem(item, node)
			if index != -1 {
				// item found, remove the item and condense tree upwards
				copy(node.Children[index:], node.Children[index+1:])
				node.Children[len(node.Children)-1] = nil
				node.Children = node.Children[:len(node.Children)-1]
				path = append(path, node)
				tr.condense(path)
				goto done
			}
		}
		if !goingUp && !node.Leaf && node.contains(&bbox) { // go down
			path = append(path, node)
			indexes = append(indexes, i)
			i = 0
			parent = node
			node = node.Children[0].(*TreeNode)
		} else if parent != nil { // go right
			i++
			if i >= len(parent.Children) {
				node = nil
			} else {
				node = parent.Children[i].(*TreeNode)
			}
			goingUp = false
		} else {
			node = nil
		}
	}
done:
	tr.reusePath = path
}

func (tr *RBush) build(items []Item, left, right int, height int) *TreeNode {
	N := right - left + 1
	M := tr.maxEntries
	var node *TreeNode

	if N <= M {
		// reached leaf level; return leaf
		// items.slice(left, right+1)
		children := make([]interface{}, 0, right+1-left-1)
		for i := left; i < right+1; i++ {
			children = append(children, items[i])
		}
		node = createNode(children)
		calcBBox(node)
		return node
	}

	if height <= 0 {
		// target height of the bulk-loaded tree
		height = int(math.Ceil(math.Log(float64(N)) / math.Log(float64(M))))

		// target number of root entries to maximize storage utilization
		M = int(math.Ceil(float64(N) / math.Pow(float64(M), float64(height-1))))
	}

	node = createNode(nil)
	node.Leaf = false
	node.height = height

	// split the items into M mostly square tiles

	N2 := int(math.Ceil(float64(N) / float64(M)))
	N1 := int(float64(N2) * math.Ceil(math.Sqrt(float64(M))))

	multiSelect(items, left, right, N1, 0)

	for i := left; i <= right; i += N1 {
		right2 := int(mathMin(float64(i+N1-1), float64(right)))

		multiSelect(items, i, right2, N2, 1)

		for j := i; j <= right2; j += N2 {
			right3 := int(mathMin(float64(j+N2-1), float64(right2)))

			// pack each entry recursively
			node.Children = append(node.Children, tr.build(items, j, right3, height-1))
		}
	}

	calcBBox(node)

	return node
}

func (tr *RBush) chooseSubtree(bbox, node *TreeNode, level int, path []*TreeNode) (*TreeNode, []*TreeNode) {
	for {
		path = append(path, node)
		if node.Leaf || len(path)-1 == level {
			break
		}
		minArea := mathInfPos
		minEnlargement := mathInfPos
		var targetNode *TreeNode
		for _, ptr := range node.Children {
			child := ptr.(*TreeNode)
			area := child.area()
			enlargement := bbox.enlargedArea(child) - area
			if enlargement < minEnlargement {
				minEnlargement = enlargement
				if area < minArea {
					minArea = area
				}
				targetNode = child
			} else if enlargement == minEnlargement {
				if area < minArea {
					minArea = area
					targetNode = child
				}
			}
		}
		if targetNode != nil {
			node = targetNode
		} else if len(node.Children) > 0 {
			node = node.Children[0].(*TreeNode)
		} else {
			// node = nil
			panic("node will be nil")
		}
	}
	return node, path
}

func (tr *RBush) insertItem(item Item) {
	var bbox TreeNode
	fillBBox(item, &bbox)
	tr.insert(&bbox, item, tr.Data.height-1)
}

func (tr *RBush) insertNode(node *TreeNode, level int) {
	tr.insert(node, node, level)
}

func (tr *RBush) insert(bbox *TreeNode, item interface{}, level int) {
	tr.reusePath = tr.reusePath[:0]
	node, insertPath := tr.chooseSubtree(bbox, tr.Data, level, tr.reusePath)
	node.Children = append(node.Children, item)
	node.extend(bbox)
	for level >= 0 {
		if len(insertPath[level].Children) > tr.maxEntries {
			insertPath = tr.split(insertPath, level)
			level--
		} else {
			break
		}
	}
	tr.adjustParentBBoxes(bbox, insertPath, level)
	tr.reusePath = insertPath
}

func (tr *RBush) split(insertPath []*TreeNode, level int) []*TreeNode {
	node := insertPath[level]
	M := len(node.Children)
	m := tr.minEntries

	tr.chooseSplitAxis(node, m, M)
	splitIndex := tr.chooseSplitIndex(node, m, M)

	spliced := make([]interface{}, len(node.Children)-splitIndex)
	copy(spliced, node.Children[splitIndex:])
	node.Children = node.Children[:splitIndex]

	newNode := createNode(spliced)
	newNode.height = node.height
	newNode.Leaf = node.Leaf

	calcBBox(node)
	calcBBox(newNode)

	if level != 0 {
		insertPath[level-1].Children = append(insertPath[level-1].Children, newNode)
	} else {
		tr.splitRoot(node, newNode)
	}
	return insertPath
}

func (tr *RBush) splitRoot(node, newNode *TreeNode) {
	tr.Data = createNode([]interface{}{node, newNode})
	tr.Data.height = node.height + 1
	tr.Data.Leaf = false
	calcBBox(tr.Data)
}

func (tr *RBush) chooseSplitIndex(node *TreeNode, m, M int) int {
	index := -1
	minOverlap := mathInfPos
	minArea := mathInfPos

	for i := m; i <= M-m; i++ {
		bbox1 := distBBox(node, 0, i, nil)
		bbox2 := distBBox(node, i, M, nil)

		overlap := bbox1.intersectionArea(bbox2)
		area := bbox1.area() + bbox2.area()

		// choose distribution with minimum overlap
		if overlap < minOverlap {
			minOverlap = overlap
			index = i

			if area < minArea {
				minArea = area
			}
		} else if overlap == minOverlap {
			// otherwise choose distribution with minimum area
			if area < minArea {
				minArea = area
				index = i
			}
		}
	}
	// return index || M - m;
	if index >= 0 {
		return index
	} else {
		return M - m
	}
}

func (tr *RBush) chooseSplitAxis(node *TreeNode, m, M int) {
	xMargin := tr.allDistMargin(node, m, M, 0)
	yMargin := tr.allDistMargin(node, m, M, 1)
	if xMargin < yMargin {
		sortNodes(node, 0)
	}
}

type leafByDim struct {
	node *TreeNode
	axis int
}

func (arr *leafByDim) Len() int { return len(arr.node.Children) }
func (arr *leafByDim) Less(i, j int) bool {
	var a, b TreeNode
	fillBBox(arr.node.Children[i].(Item), &a)
	fillBBox(arr.node.Children[j].(Item), &b)
	return a.Min[arr.axis] < b.Min[arr.axis]
}

func (arr *leafByDim) Swap(i, j int) {
	arr.node.Children[i], arr.node.Children[j] = arr.node.Children[j], arr.node.Children[i]
}

type nodeByDim struct {
	node *TreeNode
	axis int
}

func (arr *nodeByDim) Len() int { return len(arr.node.Children) }
func (arr *nodeByDim) Less(i, j int) bool {
	a := arr.node.Children[i].(*TreeNode)
	b := arr.node.Children[j].(*TreeNode)
	return a.Min[arr.axis] < b.Min[arr.axis]
}

func (arr *nodeByDim) Swap(i, j int) {
	arr.node.Children[i], arr.node.Children[j] = arr.node.Children[j], arr.node.Children[i]
}

func sortNodes(node *TreeNode, axis int) {
	if node.Leaf {
		sort.Sort(&leafByDim{node: node, axis: axis})
	} else {
		sort.Sort(&nodeByDim{node: node, axis: axis})
	}
}

// allDistMargin sorts the node's children based on the their margin for
// the specified axis
func (tr *RBush) allDistMargin(node *TreeNode, m, M int, axis int) float64 {
	sortNodes(node, axis)
	leftBBox := distBBox(node, 0, m, nil)
	rightBBox := distBBox(node, M-m, M, nil)
	margin := leftBBox.margin() + rightBBox.margin()

	var i int

	if node.Leaf {
		var child TreeNode
		for i = m; i < M-m; i++ {
			fillBBox(node.Children[i].(Item), &child)
			leftBBox.extend(&child)
			margin += leftBBox.margin()
		}
		for i = M - m - 1; i >= m; i-- {
			fillBBox(node.Children[i].(Item), &child)
			leftBBox.extend(&child)
			margin += rightBBox.margin()
		}
	} else {
		for i = m; i < M-m; i++ {
			child := node.Children[i].(*TreeNode)
			leftBBox.extend(child)
			margin += leftBBox.margin()
		}
		for i = M - m - 1; i >= m; i-- {
			child := node.Children[i].(*TreeNode)
			leftBBox.extend(child)
			margin += rightBBox.margin()
		}
	}
	return margin
}

func (tr *RBush) adjustParentBBoxes(bbox *TreeNode, path []*TreeNode, level int) {
	// adjust bboxes along the given tree path
	for i := level; i >= 0; i-- {
		path[i].extend(bbox)
	}
}

func (tr *RBush) condense(path []*TreeNode) {
	// go through the path, removing empty nodes and updating bboxes
	var siblings []interface{}
	for i := len(path) - 1; i >= 0; i-- {
		if len(path[i].Children) == 0 {
			if i > 0 {
				siblings = path[i-1].Children
				index := -1
				for j := 0; j < len(siblings); j++ {
					if siblings[j] == path[i] {
						index = j
						break
					}
				}
				copy(siblings[index:], siblings[index+1:])
				siblings[len(siblings)-1] = nil
				siblings = siblings[:len(siblings)-1]
				path[i-1].Children = siblings
			} else {
				tr.Clear()
			}
		} else {
			calcBBox(path[i])
		}
	}
}

func findItem(item Item, node *TreeNode) int {
	for i := 0; i < len(node.Children); i++ {
		if node.Children[i] == item {
			return i
		}
	}
	return -1
}

func calcBBox(node *TreeNode) {
	distBBox(node, 0, len(node.Children), node)
}

func distBBox(node *TreeNode, k, p int, destNode *TreeNode) *TreeNode {
	if destNode == nil {
		destNode = createNode(nil)
	} else {
		for i := 0; i < 2; i++ {
			destNode.Min[i] = mathInfPos
			destNode.Max[i] = mathInfNeg
		}
	}

	for i := k; i < p; i++ {
		ptr := node.Children[i]
		if node.Leaf {
			var child TreeNode
			fillBBox(ptr.(Item), &child)
			destNode.extend(&child)
		} else {
			child := ptr.(*TreeNode)
			destNode.extend(child)
		}
	}
	return destNode
}

func (a *TreeNode) extend(b *TreeNode) {
	for i := 0; i < len(a.Min); i++ {
		a.Min[i] = mathMin(a.Min[i], b.Min[i])
		a.Max[i] = mathMax(a.Max[i], b.Max[i])
	}
}

func (a *TreeNode) area() float64 {
	var area float64
	for i := 0; i < len(a.Min); i++ {
		if i == 0 {
			area = a.Max[i] - a.Min[i]
		} else {
			area *= a.Max[i] - a.Min[i]
		}
	}
	return area
}

func (a *TreeNode) margin() float64 {
	var area float64
	for i := 0; i < len(a.Min); i++ {
		if i == 0 {
			area = a.Max[i] - a.Min[i]
		} else {
			area += a.Max[i] - a.Min[i]
		}
	}
	return area
}

func (a *TreeNode) enlargedArea(b *TreeNode) float64 {
	var area float64
	for i := 0; i < len(a.Min); i++ {
		if i == 0 {
			area = mathMax(b.Max[i], a.Max[i]) - mathMin(b.Min[i], a.Min[i])
		} else {
			area *= mathMax(b.Max[i], a.Max[i]) - mathMin(b.Min[i], a.Min[i])
		}
	}
	return area
}

func (a *TreeNode) intersectionArea(b *TreeNode) float64 {
	var area float64
	for i := 0; i < len(a.Min); i++ {
		min := mathMax(a.Min[i], b.Min[i])
		max := mathMin(a.Max[i], b.Max[i])
		if i == 0 {
			area = mathMax(0, max-min)
		} else {
			area *= mathMax(0, max-min)
		}
	}
	return area
}

func (a *TreeNode) contains(b *TreeNode) bool {
	for i := 0; i < len(a.Min); i++ {
		if !(a.Min[i] <= b.Min[i] && b.Max[i] <= a.Max[i]) {
			return false
		}
	}
	return true
}

func (a *TreeNode) intersects(b *TreeNode) bool {
	for i := 0; i < len(a.Min); i++ {
		if !(b.Min[i] <= a.Max[i] && b.Max[i] >= a.Min[i]) {
			return false
		}
	}
	return true
}

func createNode(children []interface{}) *TreeNode {
	n := &TreeNode{
		Children: children,
		height:   1,
		Leaf:     true,
	}
	for i := 0; i < 2; i++ {
		n.Min[i] = mathInfPos
		n.Max[i] = mathInfNeg
	}
	return n
}

func compare(i, j Item, axis int) float64 {
	var a, b TreeNode
	fillBBox(i, &a)
	fillBBox(j, &b)
	return a.Min[axis] - b.Min[axis]
}

func swap(items []Item, i, j int) {
	items[i], items[j] = items[j], items[i]
}

func quickselect(arr []Item, k, left, right int, axis int) {
	for right > left {
		if right-left > 600 {
			n := float64(right - left + 1)
			m := float64(k - left + 1)
			z := math.Log(n)
			s := 0.5 * math.Exp(2.0*z/3.0)
			var d float64
			if m-n/2 < 0 {
				d = -1
			} else {
				d = 1
			}
			sd := 0.5 * math.Sqrt(z*s*(n-s)/n) * d
			newLeft := mathMax(float64(left), math.Floor(float64(k)-m*s/n+sd))
			newRight := mathMin(float64(right), math.Floor(float64(k)+(n-m)*s/n+sd))
			quickselect(arr, k, int(newLeft), int(newRight), axis)
		}

		t := arr[k]
		i := left
		j := right

		swap(arr, left, k)
		if compare(arr[right], t, axis) > 0 {
			swap(arr, left, right)
		}

		for i < j {
			swap(arr, i, j)
			i++
			j--
			for compare(arr[i], t, axis) < 0 {
				i++
			}
			for compare(arr[j], t, axis) > 0 {
				j--
			}
		}

		if compare(arr[left], t, axis) == 0 {
			swap(arr, left, j)
		} else {
			j++
			swap(arr, j, right)
		}

		if j <= k {
			left = j + 1
		}
		if k <= j {
			right = j - 1
		}
	}
}

func multiSelect(arr []Item, left, right, n int, axis int) {
	stack := []int{left, right}

	for len(stack) > 0 {
		right = stack[len(stack)-1]
		left = stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		if right-left <= n {
			continue
		}

		mid := left + int(math.Ceil(float64(right-left)/float64(n)/2.0)*float64(n))
		quickselect(arr, mid, left, right, axis)

		stack = append(stack, left)
		stack = append(stack, mid)
		stack = append(stack, mid)
		stack = append(stack, right)
	}
}
