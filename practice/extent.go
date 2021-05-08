package practice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type wordExtent struct {
	Count int
	Last  int64
}

type pracExtent struct {
	file            string
	words           map[string]*wordExtent
	RestudyInterval map[int]int
}

func (p *pracExtent) Save() error {
	f, err := os.Create(p.file)
	if err != nil {
		return err
	}
	defer f.Close()

	words := map[string]*wordExtent{}
	for w, e := range p.words {
		w = strings.TrimSpace(w)
		w = strings.ToLower(w)
		words[w] = e
	}
	p.words = words

	b, err := json.Marshal(p.words)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

func (p *pracExtent) Load() error {
	filebyte, err := ioutil.ReadFile(p.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(filebyte, &p.words)
}

func (p *pracExtent) ReviewWords() []string {
	words := []string{}
	now := time.Now().Unix()
	rememberCount := p.rememberCount()
	for word, extent := range p.words {
		if extent.Count >= rememberCount {
			continue
		}
		for count, hour := range p.RestudyInterval {
			if count == rememberCount {
				continue
			}
			if extent.Count <= count && extent.Last+60*60*int64(hour) < now {
				words = append(words, word)
				break
			}
		}
	}
	return words
}

func (p *pracExtent) RememberWords() []string {
	words := []string{}
	rememberCount := p.rememberCount()
	for word, e := range p.words {
		if e.Count >= rememberCount {
			words = append(words, word)
		}
	}
	return words
}

func (p *pracExtent) rememberCount() int {
	rememberCount := 0
	for count, hour := range p.RestudyInterval {
		if hour == -1 {
			rememberCount = count
		}
	}
	return rememberCount
}

func (p *pracExtent) Remember(w string) error {
	e, exist := p.words[w]
	if !exist {
		e = &wordExtent{}
		p.words[w] = e
	}
	p.words[w].Count = e.Count + 1
	p.words[w].Last = time.Now().Unix()
	return p.Save()
}

func (p *pracExtent) Forget(w string) error {
	e, exist := p.words[w]
	if !exist {
		e = &wordExtent{}
		p.words[w] = e
	}
	p.words[w].Count = e.Count - 1
	if p.words[w].Count < 0 {
		p.words[w].Count = 0
	}
	p.words[w].Last = time.Now().Unix()
	return p.Save()
}

func (p *pracExtent) CorrectNum(w string) int {
	e, exist := p.words[w]
	if !exist {
		return 0
	}
	return e.Count
}

func NewPracExtent(f string, restudyInterval map[int]int) (pracExtent, error) {
	p := pracExtent{file: f, RestudyInterval: restudyInterval}
	if p.rememberCount() == 0 {
		return p, fmt.Errorf("pracExtent rememberCount can not be 0")
	}
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		err := p.Save()
		if err != nil {
			return p, err
		}
	}
	err = p.Load()
	return p, err
}
