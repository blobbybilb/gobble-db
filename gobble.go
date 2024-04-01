package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"
)

// Storage Structure:
// - DB directory
//   - Collection1 directory
//     - numbered files each containing a gob encoded struct: "d1.gob" "d2.gob" ...
//     - metadata file: "meta.gob"

type DB struct {
	Path string
}

type Collection[T any] struct {
	Name    string
	DB      DB
	Indices []Index[T, any]
}

type Index[T any, D comparable] struct {
	Collection *Collection[T]
	Index      map[D][]string
	Extractor  func(T) D
}

type Query[T any] func(T) bool
type Updater[T any] func(T) T

type CollectionMetadata[T any] struct {
	LastID int
}

func OpenDB(path string) (DB, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return DB{}, err
	}

	return DB{Path: path}, nil
}

func (t *DB) ListCollections() ([]string, error) {
	f, err := os.Open(t.Path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	return names, nil
}

func (t *DB) CollectionExists(name string) (bool, error) {
	collections, err := t.ListCollections()
	if err != nil {
		return false, err
	}

	for _, collection := range collections {
		if collection == name {
			return true, nil
		}
	}

	return false, nil
}

func (t *DB) DeleteCollection(name string) error {
	exists, err := t.CollectionExists(name)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("collection does not exist")
	}

	if err = os.RemoveAll(t.Path + "/" + name); err != nil {
		return err
	}

	return nil
}

func initializeCollection[T any](name string, db DB) error {
	exists, err := db.CollectionExists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("collection already exists")
	}

	if err := os.MkdirAll(db.Path+"/"+name, 0755); err != nil {
		return err
	}

	file, err := os.Create(db.Path + "/" + name + "/meta.gob")
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	enc := gob.NewEncoder(file)
	err = enc.Encode(CollectionMetadata[T]{LastID: 0})
	if err != nil {
		return err
	}

	return nil
}

func OpenCollection[T any](db DB, name string) (Collection[T], error) {
	if !isValidCollectionName(name) {
		return Collection[T]{}, fmt.Errorf("invalid collection name")
	}

	exists, err := db.CollectionExists(name)
	if err != nil {
		return Collection[T]{}, err
	}

	if !exists {
		if err := initializeCollection[T](name, db); err != nil {
			return Collection[T]{}, err
		}
	}

	return Collection[T]{Name: name, DB: db}, nil
}

func (t *Collection[T]) getMetadata() (CollectionMetadata[T], error) {
	file, err := os.Open(t.DB.Path + "/" + t.Name + "/meta.gob")
	if err != nil {
		return CollectionMetadata[T]{}, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var meta CollectionMetadata[T]
	dec := gob.NewDecoder(file)
	err = dec.Decode(&meta)
	if err != nil {
		return CollectionMetadata[T]{}, err
	}

	return meta, nil
}

func (t *Collection[T]) incrementID() error {
	meta, err := t.getMetadata()
	if err != nil {
		return err
	}

	meta.LastID++

	file, err := os.Create(t.DB.Path + "/" + t.Name + "/meta.gob")
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	enc := gob.NewEncoder(file)
	err = enc.Encode(meta)
	if err != nil {
		return err
	}

	return nil
}

func (t *Collection[T]) Insert(data T) error {
	if err := t.incrementID(); err != nil {
		return err
	}

	meta, err := t.getMetadata()
	if err != nil {
		return err
	}

	file, err := os.Create(t.DB.Path + "/" + t.Name + "/d" + fmt.Sprintf("%d", meta.LastID) + ".gob")

	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	enc := gob.NewEncoder(file)
	err = enc.Encode(data)
	if err != nil {
		return err
	}

	for _, indexInterface := range t.Indices {
		index := indexInterface
		key := index.Extractor(data)
		index.Index[key] = append(index.Index[key], fmt.Sprintf("%d", meta.LastID))
	}

	return nil
}

func (t *Collection[T]) Update(query Query[T], updater Updater[T]) error {
	dir, err := os.Open(t.DB.Path + "/" + t.Name)
	if err != nil {
		return err
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(dir)

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name()[0] != 'd' {
			continue
		}

		f, err := os.Open(t.DB.Path + "/" + t.Name + "/" + file.Name())
		if err != nil {
			return err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		err = f.Close()
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		if query(data) {
			// Remove the old data from the indices
			for _, index := range t.Indices {
				key := index.Extractor(data)
				fileIDs := index.Index[key]
				for i, id := range fileIDs {
					if id == file.Name()[1:len(file.Name())-4] { // Remove the "d" prefix and ".gob" suffix
						// Remove the file id from the index
						index.Index[key] = append(fileIDs[:i], fileIDs[i+1:]...)
						break
					}
				}
			}

			data = updater(data)

			f, err := os.Create(t.DB.Path + "/" + t.Name + "/" + file.Name())
			if err != nil {
				return err
			}

			enc := gob.NewEncoder(f)
			err = enc.Encode(data)
			err = f.Close()
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}

			// Add the updated data to the indices
			for _, index := range t.Indices {
				key := index.Extractor(data)
				index.Index[key] = append(index.Index[key], file.Name()[1:len(file.Name())-4]) // Add the "d" prefix and ".gob" suffix
			}
		}
	}

	return nil
}

func (t *Collection[T]) Delete(query Query[T]) error {
	dir, err := os.Open(t.DB.Path + "/" + t.Name)
	if err != nil {
		return err
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(dir)

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name()[0] != 'd' {
			continue
		}

		f, err := os.Open(t.DB.Path + "/" + t.Name + "/" + file.Name())
		if err != nil {
			return err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		err = f.Close()
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		if query(data) {
			err = os.Remove(t.DB.Path + "/" + t.Name + "/" + file.Name())
			if err != nil {
				return err
			}

			// Update indices
			for _, index := range t.Indices {
				key := index.Extractor(data)
				fileIDs := index.Index[key]
				for i, id := range fileIDs {
					if id == file.Name()[1:len(file.Name())-4] { // Remove the "d" prefix and ".gob" suffix
						// Remove the file id from the index
						index.Index[key] = append(fileIDs[:i], fileIDs[i+1:]...)
						break
					}
				}
			}
		}
	}

	return nil
}

func (t *Collection[T]) Select(query Query[T]) ([]T, error) {
	dir, err := os.Open(t.DB.Path + "/" + t.Name)
	if err != nil {
		return nil, err
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(dir)

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var results []T
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name()[0] != 'd' {
			continue
		}

		f, err := os.Open(t.DB.Path + "/" + t.Name + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		err = f.Close()
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		if query(data) {
			results = append(results, data)
		}
	}

	return results, nil
}

func (t *Collection[T]) Count() (int, error) {
	dir, err := os.Open(t.DB.Path + "/" + t.Name)
	if err != nil {
		return 0, err
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(dir)

	files, err := dir.Readdir(-1)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "d") {
			count++
		}
	}

	return count, nil
}

func BuildIndex[T any, D comparable](c *Collection[T], extractor func(T) D) (map[D][]string, error) {
	// Open the directory of the collection
	dir, err := os.Open(c.DB.Path + "/" + c.Name)
	if err != nil {
		return nil, err
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(dir)

	// Read all the files in the directory
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Initialize the index
	index := make(map[D][]string)

	// Iterate over the files
	for _, file := range files {
		// Skip directories and non-data files
		if file.IsDir() || file.Name()[0] != 'd' {
			continue
		}

		// Open the file
		f, err := os.Open(c.DB.Path + "/" + c.Name + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		// Decode the data from the file
		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		if err != nil {
			return nil, err
		}
		err = f.Close()
		if err != nil {
			return nil, err
		}

		// Extract the key from the data
		key := extractor(data)

		// Add the file id to the index
		fileID := file.Name()[1 : len(file.Name())-4] // Remove the "d" prefix and ".gob" suffix
		index[key] = append(index[key], fileID)
	}

	return index, nil
}

func OpenIndex[T any, D comparable](c *Collection[T], extractor func(T) D) (Index[T, any], error) {
	index, err := BuildIndex(c, extractor)
	if err != nil {
		return Index[T, any]{}, err
	}

	x := func(a T) any {
		return extractor(a)
	}

	y := map[any][]string{}

	for k, v := range index {
		y[k] = make([]string, len(v))
		for i, id := range v {
			y[k][i] = id
		}
	}

	indexInterface := Index[T, any]{Index: y, Extractor: x, Collection: c}

	c.Indices = append(c.Indices, indexInterface)

	return indexInterface, nil
}

func (t *Index[T, D]) Get(key D) ([]T, error) {
	// Check if the key exists in the index
	fileIDs, ok := t.Index[key]
	if !ok {
		// If the key does not exist, return an empty slice and no error
		return []T{}, nil
	}

	// Initialize a slice to hold the results
	var results []T

	// Iterate over the file IDs associated with the key
	for _, fileID := range fileIDs {
		// Open the file
		f, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return nil, err
		}

		// Decode the data from the file
		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		if err != nil {
			_ = f.Close()
			return nil, err
		}

		// Close the file
		err = f.Close()
		if err != nil {
			return nil, err
		}

		// Append the data to the results
		results = append(results, data)
	}

	// Return the results and no error
	return results, nil
}

func (t *Index[T, D]) Del(key D) error {
	fileIDs, ok := t.Index[key]
	if !ok {
		return nil
	}

	for _, fileID := range fileIDs {
		err := os.Remove(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return err
		}
	}

	delete(t.Index, key)

	return nil
}

//func (t *Collection[T]) ListIndexes() ([]string, error) {
//	dir, err := os.Open(t.DB.Path + "/" + t.Name)
//	if err != nil {
//		return nil, err
//	}
//	defer func(dir *os.File) {
//		err := dir.Close()
//		if err != nil {
//			fmt.Println(err)
//		}
//	}(dir)
//
//	files, err := dir.Readdir(-1)
//	if err != nil {
//		return nil, err
//	}
//
//	var indexes []string
//	for _, file := range files {
//		if file.IsDir() {
//			continue
//		}
//
//		if file.Name()[0] != 'i' {
//			continue
//		}
//
//		indexes = append(indexes, file.Name()[1:len(file.Name())-4])
//	}
//
//	return indexes, nil
//}
//
//func (t *Collection[T]) IndexExists(name string) (bool, error) {
//	indexes, err := t.ListIndexes()
//	if err != nil {
//		return false, err
//	}
//
//	for _, index := range indexes {
//		if index == name {
//			return true, nil
//		}
//	}
//
//	return false, nil
//}
//
//func (t *Collection[T]) DeleteIndex(name string) error {
//	err := os.Remove(t.DB.Path + "/" + t.Name + "/i" + name + ".gob")
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//

//
//func OpenIndex[T any, D comparable](collection Collection[T], name string, extractor func(T) D) (Index[T, D], error) {
//	if !isValidIndexName(name) {
//		return Index[T, D]{}, fmt.Errorf("invalid index name")
//	}
//	index := Index[T, D]{Collection: collection, Extractor: extractor, Name: name}
//
//	exists, err := collection.IndexExists(name)
//	if err != nil {
//		return Index[T, D]{}, err
//	}
//
//	if !exists {
//		if err := index.Build(); err != nil {
//			return Index[T, D]{}, err
//		}
//	}
//
//	return index, nil
//}
//
//func (t *Index[T, D]) Get(key D) ([]T, error) {
//	file, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/i" + t.Name + ".gob")
//	if err != nil {
//		return nil, err
//	}
//
//	defer func(file *os.File) {
//		err := file.Close()
//		if err != nil {
//			fmt.Println(err)
//		}
//	}(file)
//
//	var index IndexData[D]
//	dec := gob.NewDecoder(file)
//	err = dec.Decode(&index)
//	if err != nil {
//		return nil, err
//	}
//
//	ids := index[key]
//	if ids == nil {
//		return []T{}, nil
//	}
//
//	var results []T
//	for _, id := range ids {
//		f, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + id + ".gob")
//		if err != nil {
//			return nil, err
//		}
//
//		var data T
//		dec := gob.NewDecoder(f)
//		err = dec.Decode(&data)
//		if err != nil {
//			return nil, err
//		}
//		err = f.Close()
//		if err != nil {
//			return nil, err
//		}
//
//		results = append(results, data)
//	}
//
//	return results, nil
//}
