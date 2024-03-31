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
	// Benchmarks
	timing("Start")

	// Initialize the DB
	db, _ := OpenDB("testdb")
	_ = db.DeleteCollection("testcollection")
	collection, _ := OpenCollection[Person](db, "testcollection")

	timing("DB Init")

	// Insert 1000 records
	for i := 0; i < 5; i++ {
		_ = collection.Insert(Person{Name: fmt.Sprintf("Person %d", i), Age: i})
	}

	timing("Insert 1000 records")

	index, err := OpenIndex(&collection, func(p Person) string {
		return p.Name
	})
	if err != nil {
		return
	}

	timing("Index Init")
	fmt.Println(index.Get("Person 956"))
	timing("Get 1 record from index")

	_ = collection.Insert(Person{Name: "Someone", Age: 10000})
	timing("Insert 1 record")

	x, e := index.Get("Someone")

	fmt.Println(1111, x, e)

	fmt.Println(collection.Select(func(p Person) bool { return p.Name == "Person 956" }))
	timing("Get 1 record from collection")

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
	results, _ := collection.Select(
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
