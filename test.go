package main

import "fmt"

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
		return
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

func main() {
	test()
	bench()
}
