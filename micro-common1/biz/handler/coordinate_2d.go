package handler

import "math"

type Point struct {
	X, Y float64
}

//两个点之间的距离
func (p Point) Distance(p2 Point) float64 {
	return math.Sqrt(math.Pow(p.X-p2.X, 2) + math.Pow(p.Y-p2.Y, 2))
}

type Polygon struct {
	points []Point
}

func NewPolygon(points []Point) *Polygon {
	return &Polygon{
		points: points,
	}
}

func (p *Polygon) Reset() {
	p.points = p.points[:0]
}

func (p *Polygon) Points() []Point {
	return p.points
}

func (p *Polygon) Add(point *Point) {
	for i := 0;i<len(p.points);i++{
		//已经存在的点，不增加
		if math.Abs(p.points[i].X-point.X) <= 0.001 &&
			math.Abs(p.points[i].Y-point.Y) <= 0.001{
			return
		}
	}
	p.points = append(p.points, *point)
}

func (p *Polygon) IsClosed() bool {
	return len(p.points) >= 3
}

func (p *Polygon) Contains(point *Point) bool {
	if !p.IsClosed() {
		return false
	}
	start := len(p.points) - 1
	end := 0
	contains := p.intersectsWithRaycast(*point, p.points[start], p.points[end])
	for i := 1; i < len(p.points); i++ {
		if p.intersectsWithRaycast(*point, p.points[i-1], p.points[i]) {
			contains = !contains
		}
	}
	return contains
}

// Using the raycast algorithm, this returns whether or not the passed in point
// Intersects with the edge drawn by the passed in start and end points.
// Original implementation: http://rosettacode.org/wiki/Ray-casting_algorithm#Go
func (p *Polygon) intersectsWithRaycast(point Point, start Point, end Point) bool {
	// Always ensure that the the first point
	// has a y coordinate that is less than the second point
	if start.Y > end.Y {

		// Switch the points if otherwise.
		start, end = end, start

	}

	// Move the point's y coordinate
	// outside of the bounds of the testing region
	// so we can start drawing a ray
	for point.Y == start.Y || point.Y == end.Y {
		newY := math.Nextafter(point.Y, math.Inf(1))
		point = Point{point.X, newY}
	}

	// If we are outside of the polygon, indicate so.
	if point.Y < start.Y || point.Y > end.Y {
		return false
	}

	if start.X > end.X {
		if point.X > start.X {
			return false
		}
		if point.X < end.X {
			return true
		}

	} else {
		if point.X > end.X {
			return false
		}
		if point.X < start.X {
			return true
		}
	}

	raySlope := (point.Y - start.Y) / (point.X - start.X)
	diagSlope := (end.Y - start.Y) / (end.X - start.X)

	return raySlope >= diagSlope
}

//功能同Contains，参照算法
//https://www.cnblogs.com/anningwang/p/7581545.html
func (p *Polygon) PtInPolygon(pt *Point) bool {
	if !p.IsClosed() {
		return false
	}
	min := Point{math.MaxInt32, math.MaxInt32}
	max := Point{math.MinInt32, math.MinInt32}
	for i := 0; i < len(p.points); i++ {
		if min.X > p.points[i].X {
			min.X = p.points[i].X
		}
		if min.Y > p.points[i].Y {
			min.Y = p.points[i].Y
		}
		if max.X < p.points[i].X {
			max.X = p.points[i].X
		}
		if max.Y < p.points[i].Y {
			max.Y = p.points[i].Y
		}
	}
	//在这个区域的最大平面区域之外，直接返回
	if pt.X < min.X || pt.X > max.X ||
		pt.Y < min.Y || pt.Y > max.Y {
		return false
	}
	nlen := len(p.points)
	j := nlen - 1
	result := false
	for i := 0; i < nlen; i++ {
		if ((p.points[i].Y > pt.Y) != (p.points[j].Y > pt.Y)) &&
			(pt.X < (p.points[j].X-p.points[i].X)*(pt.Y-p.points[i].Y)/(p.points[j].Y-p.points[i].Y)+p.points[i].X) {
			result = !result
		}
		j = i
	}
	return result
}
