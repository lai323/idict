package wordset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/afero"
)

type WordCache struct {
	StorageDir string
}

func NewWordCache(dir string) (WordCache, error) {
	wordcache := WordCache{StorageDir: dir}
	err := afero.NewOsFs().MkdirAll(wordcache.CacheDir(), 0755)
	if err != nil {
		return wordcache, fmt.Errorf("WordCache MkdirAll %s", err.Error())
	}
	return wordcache, nil
}

func (c WordCache) CacheDir() string {
	return fmt.Sprintf("%s/wordcache", c.StorageDir)
}

func (c WordCache) Get(text string) (Word, bool, error) {
	var err error
	var word Word

	file := path.Join(c.CacheDir(), text)
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return word, false, nil
	}
	if err != nil {
		return word, false, fmt.Errorf("WordCache Get %s", err.Error())
	}

	filebyte, err := ioutil.ReadFile(file)
	if err != nil {
		return word, false, fmt.Errorf("WordCache read file %s %s", file, err.Error())
	}
	err = json.Unmarshal(filebyte, &word)
	if err != nil {
		return word, false, fmt.Errorf("WordCache Get json Unmarshal %s %s", file, err.Error())
	}

	return word, true, nil
}

func (c WordCache) Set(word Word) error {
	file := path.Join(c.CacheDir(), word.Text)
	filebyte, err := json.Marshal(word)
	if err != nil {
		return fmt.Errorf("WordCache Set json Unmarshal %s %s", file, err.Error())
	}
	err = ioutil.WriteFile(file, filebyte, 0644)
	if err != nil {
		return fmt.Errorf("WordCache WriteFile %s %s", file, err.Error())
	}
	return nil
}
