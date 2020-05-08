package db

import (
	"bytes"
	"encoding/gob"
	"log"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/nilsocket/def/pkg/vocab"
)

var db *badger.DB

// ErrKeyNotFound is returned when key isn't found
var ErrKeyNotFound = badger.ErrKeyNotFound

// Open db
func Open(path string) {
	var err error
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err = badger.Open(opts)
	if err != nil {
		log.Fatalln(err)
	}
}

// Get key
func Get(key string) (*vocab.Word, error) {
	word := &vocab.Word{}

	err := db.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			// decode
			err = gob.NewDecoder(bytes.NewReader(val)).Decode(word)

			if err != nil {
				log.Println(err)
				return err
			}

			return nil
		})

	})

	if err != nil && err == badger.ErrKeyNotFound {
		return nil, ErrKeyNotFound
	} else if err != nil {
		return nil, err
	}

	return word, nil
}

// Put key and val into db
func Put(key string, val *vocab.Word) error {
	b := &strings.Builder{}

	// encode
	err := gob.NewEncoder(b).Encode(val)
	if err != nil {
		log.Println(err)
		return err
	}

	// Update
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(b.String()))
		return err
	})

	if err != nil {
		log.Println(err)
	}

	return err
}

// Del key from db
func Del(key string) error {
	return db.Update(func(txn *badger.Txn) error {
		txn.Delete([]byte(key))
		return nil
	})
}

// Iterate over all key value pairs and execute fn for each key
func Iterate(fn func(key string)) {
	db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{}
		// opts.PrefetchValues = true
		it := txn.NewIterator(opts)

		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().Key()
			fn(string(key))
		}

		it.Close()

		return nil
	})
}

// Close db
func Close() {
	db.Close()
}
