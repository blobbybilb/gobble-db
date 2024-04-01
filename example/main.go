package main

import (
	"fmt"

	"github.com/blobbybilb/gobble-db"
)

// Define your struct
type Shape struct {
	Name        string
	NumSides    int
	SideLengths *[]int
}

func main() {
	db, _ := gobble.OpenDB("test-db")

	// Pass in your struct as a type parameter
	shapes, _ := gobble.OpenCollection[Shape](db, "shapes")

	// Start inserting data
	shapes.Insert(Shape{"Square", 4, &[]int{4, 4, 4, 2}})

	// To query data, pass Select a function that takes in your struct and returns a boolean <- that's a "query" function
	// query: func(Shape) bool -> ([]Shape, error)
	result, _ := shapes.Select(func(shape Shape) bool { return shape.NumSides == 4 })
	fmt.Println(result) // result is a slice of Shape structs

	// To update data, pass Modify a query function,  and a function that takes in your struct and returns a modified struct <- that's an "updater" function
	// query: func(Shape) bool, updater: func(Shape) Shape
	shapes.Modify(
		func(shape Shape) bool { return shape.NumSides == 4 },
		func(shape *Shape) *Shape { shape.SideLengths[3] = 2; return shape })

	// To delete data, pass Delete a function that takes in your struct and returns a boolean
	// query: func(Shape) bool
	shapes.Delete(func(shape Shape) bool { return shape.NumSides > 10 })

	// Indexing
	// Indexing is done by passing in a function that takes in your struct and returns a value to index on <- that's an "extractor" function
	// The first type parameter is the type of the struct, and the second type parameter is the type of the index (in this case, string)
	// collection: *Collection[Shape], extractor: func(Shape) string -> (Index[Shape, string], error)
	nameIndex, _ := gobble.OpenIndex[Shape, string](&shapes, func(shape Shape) string { return shape.Name })

	// Now you can query the index by passing Get a value of the type that your extractor function returns (in this case, string)
	// key: string -> ([]Shape, error)
	fmt.Println(nameIndex.Get("Square"))

	// You can do more with indexes
	// This index has a key type of int (that represents the sum of the side lengths of the shape)
	// collection: *Collection[Shape], extractor: func(Shape) int -> (Index[Shape, int], error)
	perimeterIndex, _ := gobble.OpenIndex[Shape, int](&shapes, func(u Shape) int {
		sum := 0
		for _, l := range u.SideLengths {
			sum += l
		}
		return sum
	})

	// You can query it the same way, but with an int key
	perimeterIndex.Get(16)

	// Other functions
	shapes.Number()         // Returns the number of elements in the collection
	nameIndex.Num("Square") // Returns the number of matching elements
	nameIndex.Del("Square") // Deletes the index for the key "Square"

}
