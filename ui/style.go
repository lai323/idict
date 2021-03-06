package ui

import (
	te "github.com/muesli/termenv"
)

var (
	StyleLogo = NewStyle("#ffc27d", "#f37329", true)
	StyleHelp = NewStyle("#4e4e4e", "", true)
)

func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(te.ColorProfile().Color(fg)).Background(te.ColorProfile().Color(bg))
	if bold {
		s = s.Bold()
	}
	return s.Styled
}
