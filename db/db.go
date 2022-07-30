package db

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"
)

const (
	COLL_BUCKET  = "collection"
	PRAC_BUCKET  = "practice"
	DICT_BUCKET  = "dict"
	DEFAULT_COLL = "default"
)

type DictDB interface {
	Close() error
	Put(word Word) error
	Get(text string) (*Word, error)
	CollectionPut(collection string, word []string) error
	CollectionList() (map[string]int, error)
	CollectionWords(collection string) ([]string, error)
	CollectionDelete(collection string) error
	PracticePut(WordPractice) error
	PracticeGet(text string) (WordPractice, error)
	Practice() (map[string]WordPractice, error)
	PracticeClean() error
}

type BoltDictDB struct {
	*bolt.DB
}

func NewBoltDictDB(path string) (*BoltDictDB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(COLL_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(DICT_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PRAC_BUCKET))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &BoltDictDB{DB: db}, nil
}

func (db *BoltDictDB) Put(word Word) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := gob.NewEncoder(buf).Encode(word)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(DICT_BUCKET)).Put([]byte(word.Text), buf.Bytes())
	})
}

func (db *BoltDictDB) Get(text string) (*Word, error) {
	data := []byte{}
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DICT_BUCKET)).Get([]byte(text))
		copy(data, b)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	w := Word{}
	err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w, err
}

func (db *BoltDictDB) CollectionPut(collection string, word []string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket([]byte(COLL_BUCKET)).
			CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return err
		}
		for _, w := range word {
			err = b.Put([]byte(w), nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *BoltDictDB) CollectionList() (map[string]int, error) {
	collections := map[string]int{}
	err := db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(COLL_BUCKET))
		b.ForEach(func(k, v []byte) error {
			collections[string(k)] = b.Bucket(k).Stats().KeyN
			return nil
		})
		return nil
	})
	if err != nil {
		return collections, nil
	}
	return collections, nil
}

func (db *BoltDictDB) CollectionWords(collection string) ([]string, error) {
	words := []string{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COLL_BUCKET)).Bucket([]byte(collection))
		b.ForEach(func(k, _ []byte) error {
			words = append(words, string(k))
			return nil
		})
		return nil
	})
	return words, err
}

func (db *BoltDictDB) CollectionDelete(collection string) error {
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(COLL_BUCKET))
		return b.Delete([]byte(collection))
	})
	return err
}

func (db *BoltDictDB) PracticePut(word WordPractice) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := gob.NewEncoder(buf).Encode(word)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(PRAC_BUCKET)).Put([]byte(word.Text), buf.Bytes())
	})
}

func (db *BoltDictDB) PracticeGet(text string) (WordPractice, error) {
	var data []byte
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PRAC_BUCKET)).Get([]byte(text))
		data = append(data, b...)
		return nil
	})
	if err != nil {
		return WordPractice{Text: text}, err
	}
	if len(data) == 0 {
		return WordPractice{Text: text}, nil
	}
	w := WordPractice{}
	err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&w)
	if err != nil {
		return w, err
	}
	return w, err
}

func (db *BoltDictDB) Practice() (map[string]WordPractice, error) {
	words := map[string]WordPractice{}
	err := db.DB.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(PRAC_BUCKET)).ForEach(func(k, v []byte) error {
			w := WordPractice{}
			err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&w)
			if err != nil {
				return err
			}
			words[w.Text] = w
			return nil
		})
	})
	return words, err
}

func (db *BoltDictDB) PracticeClean() error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(PRAC_BUCKET))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(PRAC_BUCKET))
		return err
	})
}
