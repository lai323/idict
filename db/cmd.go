package db

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"time"
)

var validword = regexp.MustCompile(`^[A-Za-z-]+[A-Za-z-\.']*$`)

func CollectionImport(p, name string, db DictDB) error {
	file_name := strings.Split(path.Base(p), ".")[0]
	if file_name == "/" {
		return fmt.Errorf("invalid word set path %s ", name)
	}
	if name == "" {
		name = file_name
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("read file %s %s", p, err.Error())
	}

	words := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		word := strings.TrimSpace(line)
		if word == "" {
			continue
		}
		if !validword.MatchString(word) {
			return fmt.Errorf("invalid word '%s'", word)
		}
		words = append(words, word)
	}
	return db.CollectionPut(name, words)
}

func CollectionList(db DictDB) error {
	col, err := db.CollectionList()
	if err != nil {
		return err
	}
	for k, v := range col {
		fmt.Printf("%-30s%d\n", string(k), v)
	}
	return nil
}

func CollectionWords(name string, db DictDB) error {
	words, err := db.CollectionWords(name)
	if err != nil {
		return err
	}
	for _, w := range words {
		fmt.Println(w)
	}
	return nil
}

func CollectionDelete(name string, db DictDB) error {
	return db.CollectionDelete(name)
}

func PracticeDegree(db DictDB) error {
	words, err := db.Practice()
	if err != nil {
		return err
	}
	for _, w := range words {
		t := time.Unix(w.LastTime, 0).Format("2006-01-02 15:06")
		fmt.Printf("%-20s%-10d%s\n", w.Text, w.Degree, t)
	}
	return nil
}
