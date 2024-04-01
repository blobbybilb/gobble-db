package gobble

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
	Indices []Index[T, any] // Go doesn't seem to support generics here, this is internal so `any` is fine
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

func OpenIndex[T any, D comparable](c *Collection[T], extractor func(T) D) (Index[T, any], error) {
	index, err := buildIndex(c, extractor)
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

func buildIndex[T any, D comparable](c *Collection[T], extractor func(T) D) (map[D][]string, error) {
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

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	index := make(map[D][]string)

	for _, file := range files {
		if file.IsDir() || file.Name()[0] != 'd' {
			continue
		}

		f, err := os.Open(c.DB.Path + "/" + c.Name + "/" + file.Name())
		if err != nil {
			return nil, err
		}

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

		key := extractor(data)

		fileID := file.Name()[1 : len(file.Name())-4]
		index[key] = append(index[key], fileID)
	}

	return index, nil
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

func (t *Collection[T]) Modify(query Query[T], updater Updater[T]) error {
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
					if id == file.Name()[1:len(file.Name())-4] {
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
				index.Index[key] = append(index.Index[key], file.Name()[1:len(file.Name())-4])
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

			// Modify indices
			for _, index := range t.Indices {
				key := index.Extractor(data)
				fileIDs := index.Index[key]
				for i, id := range fileIDs {
					if id == file.Name()[1:len(file.Name())-4] {
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

func (t *Collection[T]) Number() (int, error) {
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

func (t *Index[T, D]) Get(key D) ([]T, error) {
	fileIDs, ok := t.Index[key]
	if !ok {
		// If the key does not exist, return an empty slice and no error
		return []T{}, nil
	}

	var results []T

	for _, fileID := range fileIDs {
		f, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return nil, err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		if err != nil {
			_ = f.Close()
			return nil, err
		}

		err = f.Close()
		if err != nil {
			return nil, err
		}

		results = append(results, data)
	}

	return results, nil
}

func (t *Index[T, D]) Del(key D) error {
	fileIDs, ok := t.Index[key]
	if !ok {
		return nil
	}

	fileIDsCopy := make([]string, len(fileIDs))
	copy(fileIDsCopy, fileIDs)

	if len(t.Collection.Indices) == 1 {
		// only an optimization
		for _, fileID := range fileIDsCopy {
			err := os.Remove(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
			if err != nil {
				return err
			}
		}

		delete(t.Index, key)
		return nil
	}

	for _, fileID := range fileIDsCopy {
		f, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		if err != nil {
			_ = f.Close()
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

		// Remove from indices
		for _, index := range t.Collection.Indices {
			indexKey := index.Extractor(data)
			indexFileIDs := index.Index[indexKey]
			for i, id := range indexFileIDs {
				if id == fileID {
					index.Index[indexKey] = append(indexFileIDs[:i], indexFileIDs[i+1:]...)
					break
				}
			}
		}

		err = os.Remove(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return err
		}
	}

	delete(t.Index, key)

	return nil
}

func (t *Index[T, D]) Mod(key D, updater Updater[T]) error {
	fileIDs, ok := t.Index[key]
	if !ok {
		return nil
	}

	fileIDsCopy := make([]string, len(fileIDs))
	copy(fileIDsCopy, fileIDs)

	for _, fileID := range fileIDsCopy {
		f, err := os.Open(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return err
		}

		var data T
		dec := gob.NewDecoder(f)
		err = dec.Decode(&data)
		if err != nil {
			_ = f.Close()
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

		// Remove the data from the indices
		for _, index := range t.Collection.Indices {
			indexKey := index.Extractor(data)
			indexFileIDs := index.Index[indexKey]
			for i, id := range indexFileIDs {
				if id == fileID {
					index.Index[indexKey] = append(indexFileIDs[:i], indexFileIDs[i+1:]...)
					break
				}
			}
		}

		data = updater(data)

		f, err = os.Create(t.Collection.DB.Path + "/" + t.Collection.Name + "/d" + fileID + ".gob")
		if err != nil {
			return err
		}

		enc := gob.NewEncoder(f)
		err = enc.Encode(data)
		if err != nil {
			_ = f.Close()
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

		// Add the updated data to the indices
		for _, index := range t.Collection.Indices {
			indexKey := index.Extractor(data)
			index.Index[indexKey] = append(index.Index[indexKey], fileID)
		}
	}

	return nil
}

func (t *Index[T, D]) Num(key D) (int, error) {
	fileIDs, ok := t.Index[key]
	if !ok {
		return 0, nil
	}

	return len(fileIDs), nil
}
