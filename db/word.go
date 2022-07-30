package db

type Translate struct {
	Part string
	Mean string
}

type Pronounce struct {
	Phonetic string
	Voice    string
}

type GuessWord struct {
	Label string
	Value string
	// Hlword     string
}

type Phrase struct {
	Word  string
	Text  string
	Trans string
}

type Sentence struct {
	Word  string
	Text  string
	Trans string
}

type Word struct {
	PronounceUS Pronounce
	PronounceUK Pronounce
	Text        string
	Translates  []Translate
	Phrases     []Phrase
	Sentences   []Sentence
}

type WordPractice struct {
	Text     string
	Degree   int
	LastTime int64
}
