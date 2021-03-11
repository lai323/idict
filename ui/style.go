package ui

import (
	te "github.com/muesli/termenv"
)

var (
	StyleLogo            = NewStyle("#ffc27d", "#f37329", true, false)
	StyleHelp            = NewStyle("#4e4e4e", "", true, false)
	StyleMean            = NewStyle("#ffffff", "", false, false)
	StylePart            = NewStyle("#66C2CD", "", false, true)
	StylePhrasesText     = NewStyle("#B9BFCA", "", false, false)
	StyleSentencesText   = NewStyle("#B9BFCA", "", false, false)
	StyleGuessText       = NewStyle("#D290E4", "", false, false)
	StyleGuessTextSelect = NewStyle("#ff5faf", "", true, false)
)

const (
	InputTextColor = "#ff5faf"
)

func NewStyle(fg string, bg string, bold bool, italic bool) func(string) string {
	s := te.Style{}.Foreground(te.ColorProfile().Color(fg)).Background(te.ColorProfile().Color(bg))
	if bold {
		s = s.Bold()
	}
	if italic {
		s = s.Italic()
	}
	return s.Styled
}
