package ui

import (
	"fmt"
	"strings"

	"github.com/lai323/idict/db"
)

type WordMsg struct {
	db.Word
}

type TransModel struct {
	Word db.Word
}

func (m TransModel) View() string {
	var (
		pronounce    string
		transtext    string
		phrasetext   string
		sentencetext string
	)

	if m.Word.PronounceUS.Phonetic != "" {
		pronounce += "US: " + m.Word.PronounceUS.Phonetic
	}
	if m.Word.PronounceUK.Phonetic != "" {
		pronounce += "      UK: " + m.Word.PronounceUK.Phonetic
	}
	if pronounce != "" {
		pronounce = "\n" + pronounce + "\n"
	}

	for _, t := range m.Word.Translates {
		transtext += fmt.Sprintf("%s %s\n", StyleMean(t.Mean), StylePart(t.Part))
	}
	for _, p := range m.Word.Phrases {
		phrasetext += fmt.Sprintf("%s: %s\n", StylePhrasesText(p.Text), p.Trans)
	}
	for _, s := range m.Word.Sentences {
		sentencetext += fmt.Sprintf("%s: \n    %s\n", StyleSentencesText(s.Text), s.Trans)
	}

	return strings.Join([]string{
		pronounce,
		transtext,
		phrasetext,
		sentencetext,
	}, "\n")
}

type HelpMsg struct {
}

type HelpModel struct {
	Keyhelp [][]string
	Active  bool
}

func (m HelpModel) View() string {
	var text []string
	text = append(text, "")
	text = append(text, "")
	for _, info := range m.Keyhelp {
		k, help := info[0], info[1]
		text = append(text,
			Line(
				40,
				Cell{
					Width: 4,
				},
				Cell{
					Width: 6,
					Align: LeftAlign,
					Text:  StyleKey(k),
				},
				Cell{
					Align: LeftAlign,
					Text:  StyleKeyHelp(help),
				},
			))
	}
	return strings.Join(text, "\n")
}
