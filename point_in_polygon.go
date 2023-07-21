package concaveman

func PointInPolygonOffset(point Point, poly []Point, start, end int) bool {
	x := point[0]
	y := point[1]
	inside := false
	len := end - start
	j := len - 1
	for i := 0; i < len; i++ {
		xi := poly[i+start][0]
		yi := poly[i+start][1]
		xj := poly[j+start][0]
		yj := poly[j+start][1]
		intersect := ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

func PointInPolygon(point Point, poly []Point) bool {
	return PointInPolygonOffset(point, poly, 0, len(poly))
}
