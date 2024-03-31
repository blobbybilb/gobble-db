package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

type Person struct {
	Name string
	Age  int
}

func test() {
	db, err := OpenDB("db")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = db.DeleteCollection("people")
	if err != nil {
		fmt.Println(err)
	}

	collection, err := OpenCollection[Person](db, "people")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = collection.Insert(Person{Name: "Alice", Age: 30})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = collection.Insert(Person{Name: "Bob", Age: 40})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = collection.Update(func(p Person) bool {
		return p.Age > 35
	}, func(p Person) Person {
		p.Age += 1
		return p
	})

	results, err := collection.Select(func(p Person) bool {
		fmt.Println(p)
		return p.Age > 25
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, result := range results {
		fmt.Println(result)
	}
}

//func testIndexing() {
//	db, _ := OpenDB("testdb")
//	_ = db.DeleteCollection("people")
//	collection, _ := OpenCollection[Person](db, "people")
//
//	_ = collection.Insert(Person{Name: "Alice", Age: 30})
//	_ = collection.Insert(Person{Name: "Alice", Age: 31})
//	_ = collection.Insert(Person{Name: "Bob", Age: 40})
//
//	index, _ := OpenIndex[Person, string](collection, "age", func(p Person) string {
//		return p.Name
//	})
//
//	results, _ := index.Get("Alice")
//	fmt.Println(results)
//}

//func benchIndexing() {
//	db, _ := OpenDB("testdb")
//	collection, err := OpenCollection[Person](db, "index-bench")
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	timing("indexing benchmark: DB Init")
//
//	for i := 0; i < 1000; i++ {
//		_ = collection.Insert(Person{Name: fmt.Sprintf("Person %d", i), Age: i})
//	}
//
//	timing("Insert 1000 records")
//
//	_ = collection.DeleteIndex("name")
//	index, err := OpenIndex[Person, string](collection, "name", func(p Person) string {
//		return p.Name
//	})
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	timing("Index Init")
//
//	for i := 0; i < 1000; i++ {
//		_ = collection.Insert(Person{Name: fmt.Sprintf("Person %d 2", i), Age: i})
//	}
//
//	timing("Insert 1000 records")
//
//	for i := 0; i < 500; i++ {
//		_, _ = index.Get(fmt.Sprintf("Person %d", i))
//	}
//
//	timing("Get 500 records")
//
//	for i := 0; i < 500; i++ {
//		_ = collection.Delete(func(p Person) bool {
//			return p.Age%2 == 0
//		})
//	}
//
//	timing("Delete 500 records")
//
//	collection.DeleteIndex("name")
//
//	timing("Delete Index")
//
//	_ = db.DeleteCollection("index-bench")
//}

func main() {
	//test()
	bench()
	//testIndexing()
	//benchIndexing()

	//	encode to gob
	file, _ := os.Create("test.gob")
	enc := gob.NewEncoder(file)
	_ = enc.Encode("a")
	_ = enc.Encode("b")
	_ = file.Close()

	_ = os.RemoveAll("db")
	_ = os.RemoveAll("testdb")
}
