package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

// Storage Structure:
// - DB directory
//   - Collection1 directory
//     - numbered files each containing a gob encoded struct: "d1.gob" "d2.gob" ...
//     - metadata file: "meta.gob"

type DB struct {
	Path string
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

type Collection[T any] struct {
	Name string
	DB   DB
}

func initializeCollection[T any](name string, db DB) error {
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

type Query[T any] func(T) bool
type Updater[T any] func(T) T

type CollectionMetadata[T any] struct {
	LastID int
}

func (t *Collection[T]) GetMetadata() (CollectionMetadata[T], error) {
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
	meta, err := t.GetMetadata()
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

	meta, err := t.GetMetadata()
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

// Indexing Structure:
// - DB directory
//   - Collection1 directory
//     - index files: "i<field>.gob" "i<field>.gob" ...

// Indexing implementation:
// -

//type Index[T any] struct {
//}
