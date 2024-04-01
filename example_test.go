package gobble

import (
	"fmt"
	"testing"
)

type Shape struct {
	Name        string
	NumSides    int
	SideLengths []int
}

func TestExample(t *testing.T) {
	db, _ := OpenDB("example-db")

	c, _ := OpenCollection[Shape](db, "shapes")

	// Keep an index by name
	nameIdx, _ := OpenIndex[Shape, string](&c, func(u Shape) string { return u.Name })

	// Insert some shapes
	_ = c.Insert(Shape{Name: "Triangle", NumSides: 3, SideLengths: []int{3, 4, 5}})
	_ = c.Insert(Shape{Name: "Square", NumSides: 4, SideLengths: []int{1, 1, 1, 1}})
	_ = c.Insert(Shape{Name: "Pentagon", NumSides: 5, SideLengths: []int{1, 1, 1, 1, 2}})
	_ = c.Insert(Shape{Name: "Hexagon", NumSides: 6, SideLengths: []int{1, 1, 1, 1, 1, 1}})

	// Query all shapes with more than 4 sides
	result, _ := c.Select(func(s Shape) bool { return s.NumSides > 4 })

	_ = c.Modify(func(s Shape) bool { return s.NumSides == 3 }, func(s Shape) Shape { s.SideLengths[0] = 4; return s })
	_ = c.Delete(func(s Shape) bool { return s.NumSides == 4 })

	// Get from an index
	result, _ = nameIdx.Get("Pentagon")

	// Now make an index by perimeter
	perimeterIdx, _ := OpenIndex[Shape, int](&c, func(u Shape) int {
		sum := 0
		for _, l := range u.SideLengths {
			sum += l
		}
		return sum
	})

	// Query all shapes with a perimeter of 6
	result, _ = perimeterIdx.Get(6)
	fmt.Println(result)
}
