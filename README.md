# gobble-db

A pure go (no cgo) on-disk "struct-oriented" embedded DB, with a **simple**, **friendly**, and **general** API.

#### Goal: friendliest simple embedded DB for small-to-medium-sized Go projects.

**Simple**: 0 dependencies, entire API shown in the example below, ~800 LoC implementation

**Friendly**: pure go so cross-compiling is easy, "just works" with no config, straightforward API

**General**: simple doesn't mean limited, most functions take functions/types as params to allow for flexibility without
the added complexity of a query building API, relying instead on built-in Go language features and type system
(see querying and indexing below).

`go get github.com/blobbybilb/gobble-db`

## Docs

Core API Overview:
```go
// T is the type of the struct you want to store
OpenCollection[T](db DB, name string)
collection.Insert(T)
collection.Select(func(T) bool) -> []T // function param should return true for elements you want to retrieve

// first function param should return true for elements you want
// second function param should return the modified element
collection.Modify(func(T) bool, func(T) T)

collection.Delete(func(T) bool) // function param should return true for elements you want to delete

// Indexing:

// T is the type of the struct your collection holds, K is the type of the index
// The function passed should return the value you want to index on, given a struct of type T
// This gives you the flexibility to index on any field, part of a field, a combination of fields, etc.
func OpenIndex[T, K](*Collection[T], func(T) K) -> Index[T, K]
index.Get(K) -> []T

// (Note: most of these functions also return an error type, not shown here)
```

Take a look at this example covering all the functionality:
```go
package main

import (
	"fmt"
	"github.com/blobbybilb/gobble-db"
)

// Define your data type
// Most structs containing data only would work, but they need to be serializable by the gob package
// Gobble does not require you to do any gobble-specific configuration to your data types
type Shape struct {
	Name        string
	NumSides    int
	SideLengths []int
}

func main() {
	db, _ := gobble.OpenDB("test-db")

	// Pass in your struct as a type parameter, now you have a collection of that struct
	shapes, _ := gobble.OpenCollection[Shape](db, "shapes")

	// Start inserting data
	shapes.Insert(Shape{"Square", 4, []int{4, 4, 4, 2}})

	// To query data, pass Select a function that takes in your struct and returns a boolean <-- that's a "query" function
	// query: func(Shape) bool -> ([]Shape, error)
	result, _ := shapes.Select(func(shape Shape) bool { return shape.NumSides == 4 })
	fmt.Println(result) // result is a slice of Shape structs

	// To update data, pass Modify a query function,  and a function that takes in your struct and returns a modified struct <-- that's an "updater" function
	// query: func(Shape) bool, updater: func(Shape) Shape
	shapes.Modify(
		func(shape Shape) bool { return shape.NumSides == 4 },
		func(shape Shape) Shape { shape.SideLengths[3] = 4; return shape })

	// To delete data, pass Delete a function that takes in your struct and returns a boolean
	// query: func(Shape) bool
	shapes.Delete(func(shape Shape) bool { return shape.NumSides > 10 })

	// Indexing
	// Indexing speeds up querying by storing a hash map of keys to slices of structs in-memory, without an index queries need to scan the entire collection
	// Indexing does come with a memory and a (small) write performance cost, but read performance is greatly improved
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
	nameIndex.Del("Square") // Deletes all elements that match the key
	nameIndex.Mod("Square", // Modifies all elements that match the key, takes an updater function
		func(s Shape) Shape { s.Name = "Still a Square"; return s })
}
```


## Info

### Performance

Performance is not a priority; minimal development overhead is. That said, it should be fast enough for
small-to-medium-sized projects (10k-100k items in a collection). Using some very rough benchmarks (like ~OoM):
- Inserting 10k items takes about 1-2 seconds
- Querying for ~5k of those takes about 0.5-1 seconds (not indexed)
- Querying for ~5k of those takes about 0.1-0.2 seconds (indexed)
- Querying for 1 of those takes about 0.00001-0.00003 seconds (indexed) (not indexed is about the same as for 5k not indexed)
- Modifying 5k of those takes about 1-2 seconds (not indexed)
- Deleting 5k of those takes about 0.5-1 seconds (not indexed)
- Indexing 10k items takes about 1-1.5 seconds


### Does it support transactions? Async I/O? ACID?

Nope. Too much complexity for the goal of this project.

### Is it production ready?
If your "production" use case allows you to consider a library as new and not popular as this one,
then yes, it will probably be production-ready enough for your use case. It's meant to be simple enough
that any bugs surface quickly and are easy to fix.

### Why is it called gobble-db?
It uses the Go "gob" binary data format, and it's supposed to be simple and friendly, so gobble-db.

### License
LGPLv2.1