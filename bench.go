package main

import (
	"fmt"
	"time"
)

var timeLast = time.Now().UnixMicro()

func timing(msg string) {
	timeNow := time.Now().UnixMicro()
	fmt.Printf("%s: %fs\n", msg, float64(timeNow-timeLast)/1000000)
	timeLast = timeNow
}

func bench() {
	timing("Start")

	// Initialize the DB
	db, _ := OpenDB("testdb")
	_ = db.DeleteCollection("testcollection")
	collection, _ := OpenCollection[Person](db, "testcollection")
	timing("DB Init")

	// Insert 1000 records
	for i := 0; i < 1000; i++ {
		_ = collection.Insert(Person{Name: fmt.Sprintf("Person %d", i), Age: i})
	}

	timing("Insert 1000 records (before indexing)")

	// Start indexing
	index, _ := OpenIndex[Person, string](&collection, func(p Person) string {
		return p.Name
	})

	timing("Index init")

	// Insert 1000 records
	for i := 0; i < 1000; i++ {
		_ = collection.Insert(Person{Name: fmt.Sprintf("Person %d 2", i), Age: -i})
	}

	timing("Insert 1000 records")

	// Get 500 records by index
	for i := 0; i < 500; i++ {
		_, _ = index.Get(fmt.Sprintf("Person %d", i))
	}

	timing("Get 500 records")

	// Get 500 records by query
	results, _ := collection.Select(
		func(p Person) bool {
			return 0 < p.Age && p.Age < 500
		})

	timing("Select 500 records")

	// Get 1 record by index
	_, _ = index.Get("Person 0")

	timing("Get 1 record")

	// Get 1 record by query

	results, _ = collection.Select(
		func(p Person) bool {
			return p.Name == "Person 0"
		})

	timing("Select 1 record")

	// Update 500 records
	_ = collection.Update(
		func(p Person) bool {
			return p.Age%2 == 0
		}, func(p Person) Person {
			p.Age += 1000
			return p
		})

	timing("Update 500 records")

	// Select 500 records
	results, _ = collection.Select(
		func(p Person) bool {
			return p.Age < 1000
		})

	timing("Select 500 records")
	fmt.Println(len(results), results[0], results[len(results)-1])

	// Delete 500 records
	_ = collection.Delete(
		func(p Person) bool {
			return p.Age > 1000
		})

	timing("Delete 500 records")
}
