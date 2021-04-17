package wordset

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

type WordSet struct {
	Name       string
	Words      map[string]int
	StorageDir string
}

func NewWordSet(name, dir string) (WordSet, error) {
	err := afero.NewOsFs().MkdirAll(dir, 0755)
	if err != nil {
		err = fmt.Errorf("WordSet MkdirAll %s", err.Error())
	}
	return WordSet{
		Name:       name,
		StorageDir: dir,
		Words:      map[string]int{},
	}, err
}

func (ws WordSet) fileName() string {
	return path.Join(ws.StorageDir, ws.Name+".wordset")
}

func (ws WordSet) Exist() (bool, error) {
	_, err := os.Stat(ws.fileName())
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, err
}

func (ws WordSet) Save(force bool) error {
	file := ws.fileName()
	exist, err := ws.Exist()
	if exist && !force {
		return fmt.Errorf("WordSet %s already exist", file)
	}
	if err != nil {
		return fmt.Errorf("WordSet Exist %s", err.Error())
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for word := range ws.Words {
		_, err := w.WriteString(word + "\n")
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func (ws *WordSet) Append(word string) error {
	word = strings.TrimSpace(word)
	ws.Words[word] = 0
	return ws.Save(true)
}

var validword = regexp.MustCompile(`^[a-z]+[a-z]$`)

func (ws *WordSet) Load() error {
	exist, err := ws.Exist()
	if !exist {
		return nil
	}

	file := ws.fileName()
	filebyte, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read file %s %s", file, err.Error())
	}

	for _, wordline := range strings.Split(string(filebyte), "\n") {
		word := strings.TrimSpace(wordline)
		if wordline == "" {
			continue
		}
		if !validword.MatchString(word) {
			return fmt.Errorf("WordSet Load invalid word %s", word)
		}
		ws.Words[word] = 0
	}
	return nil
}

type WordSetManage struct {
	StoragePath string
}

func (m WordSetManage) WordSetDir() string {
	return fmt.Sprintf("%s/wordset", m.StoragePath)
}

func (m WordSetManage) Import(p string) error {
	dir := path.Dir(p)
	name := strings.Split(path.Base(p), ".")[0]
	if name == "/" {
		return fmt.Errorf("invalid word set path %s ", name)
	}

	ws, err := NewWordSet(name, dir)
	if err != nil {
		return err
	}
	err = ws.Load()
	if err != nil {
		return err
	}
	ws.StorageDir = m.WordSetDir()
	return ws.Save(false)
}

func (m WordSetManage) List(p string) error {
	files, err := ioutil.ReadDir(m.WordSetDir())
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Println("    ", strings.Split(path.Base(f.Name()), ".")[0])
	}
	return nil
}

func (m WordSetManage) Show(name string) error {
	ws, err := NewWordSet(name, m.WordSetDir())
	if err != nil {
		return err
	}

	exist, err := ws.Exist()
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("WrodSet %s not exist", name)
	}
	err = ws.Load()
	if err != nil {
		return err
	}
	for word := range ws.Words {
		fmt.Println(word)
	}
	return nil
}
