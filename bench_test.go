package gobble

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const short = 10

var timeLast = time.Now().UnixMicro()

func timing(msg string) {
	timeNow := time.Now().UnixMicro()
	fmt.Printf("%s: %fs\n", msg, float64(timeNow-timeLast)/1000000)
	timeLast = timeNow
}

func TestBenchmark(t *testing.T) {
	timing("Start")

	db, _ := OpenDB("benchdb")
	_ = db.DeleteCollection("testcollection")
	collection, _ := OpenCollection[ExamplePersonStruct](db, "testcollection")
	timing("DB Init")

	for i := 0; i < 10000/short; i++ {
		_ = collection.Insert(ExamplePersonStruct{Name: fmt.Sprintf("ExamplePersonStruct %d", i), Age: i})
	}
	timing("Insert 10000 records (before indexing)")

	index, _ := OpenIndex[ExamplePersonStruct, string](&collection, func(p ExamplePersonStruct) string {
		return p.Name
	})
	timing("Index init")

	for i := 0; i < 10000/short; i++ {
		_ = collection.Insert(ExamplePersonStruct{Name: fmt.Sprintf("ExamplePersonStruct %d 2", i), Age: -i})
	}
	timing("Insert 10000 records")

	for i := 0; i < 5000/short; i++ {
		_, _ = index.Get(fmt.Sprintf("ExamplePersonStruct %d", i))
	}
	timing("Get 5000 records")

	_, _ = collection.Select(
		func(p ExamplePersonStruct) bool {
			return 0 < p.Age && p.Age < 5000/short
		})
	timing("Select 5000 records")

	_, _ = index.Get("ExamplePersonStruct 0")
	timing("Get 1 record")

	_, _ = collection.Select(
		func(p ExamplePersonStruct) bool {
			return p.Name == "ExamplePersonStruct 0"
		})
	timing("Select 1 record")

	_ = collection.Modify(
		func(p ExamplePersonStruct) bool {
			return p.Age%2 == 0
		}, func(p ExamplePersonStruct) ExamplePersonStruct {
			p.Age += 10000 / short
			return p
		})
	timing("Modify 5000 records")

	_, _ = collection.Select(
		func(p ExamplePersonStruct) bool {
			return p.Age < 10000/short
		})
	timing("Select 5000 records")
	//fmt.Println(len(results), results[0], results[len(results)-1])

	_ = collection.Delete(
		func(p ExamplePersonStruct) bool {
			return p.Age > 10000/short
		})
	timing("Delete 5000 records")

	n, _ := collection.Number()
	fmt.Println(n)

	//	benchmark index.Del and index.Mod
	for i := 0; i < 5000/short; i++ {
		_ = index.Del(fmt.Sprintf("ExamplePersonStruct %d", i))
	}
	timing("Index Del 5000 records")

	n, _ = collection.Number()
	fmt.Println(n)
	timing("Number")

	// Reset
	_ = collection.Delete(func(p ExamplePersonStruct) bool { return true })
	timing("Delete all records")

	// reinsert
	for i := 0; i < 10000/short; i++ {
		_ = collection.Insert(ExamplePersonStruct{Name: fmt.Sprintf("ExamplePersonStruct %d", i), Age: i})
	}

	// benchmark index.Mod
	for i := 0; i < 5000/short; i++ {
		_ = index.Mod(fmt.Sprintf("ExamplePersonStruct %d", i), func(p ExamplePersonStruct) ExamplePersonStruct {
			p.Age += 10000 / short
			return p
		})
	}
	timing("Index Mod 5000 records")

	//x, _ := collection.Select(func(p ExamplePersonStruct) bool { return true })
	//fmt.Println(sortPersonsByName(x))

	_ = os.RemoveAll("benchdb")
}
