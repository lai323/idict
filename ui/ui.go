package ui

import (
	"log"
	"math"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
)

func getElementWidth(widthTotal int, count int) (int, int) {
	remainder := widthTotal % count
	width := int(math.Floor(float64(widthTotal) / float64(count)))

	return width, remainder
}

type TextAlign int

const (
	LeftAlign TextAlign = iota
	RightAlign
)

func (ta TextAlign) String() string {
	return [...]string{"LeftAlign", "RightAlign"}[ta]
}

type Cell struct {
	Text  string
	Width int
	Align TextAlign
}

// 不能处理中英文夹杂的对齐情况，因为不同字符显示字宽不同
func Line(width int, cells ...Cell) string {

	widthFlex := width
	var widthFlexCells []*int

	for i, cell := range cells {
		if cell.Width <= 0 {
			widthFlexCells = append(widthFlexCells, &cells[i].Width)
			continue
		}
		widthFlex -= cell.Width
	}

	widthWithoutRemainder, remainder := getElementWidth(widthFlex, len(widthFlexCells))
	for i := range widthFlexCells {

		*widthFlexCells[i] = widthWithoutRemainder
		if i < remainder {
			*widthFlexCells[i] = widthWithoutRemainder + 1
		}
	}

	var gridLine string
	for _, cell := range cells {
		// 这种操作会把多字节的字符，例如中文分裂开，导致无法显示，像下面这样
		// fmt.Println(string([]rune("abbr. 美国政治和社会科学研究院(American ..."[:40])))
		textWidth := ansi.PrintableRuneWidth(cell.Text)
		if textWidth > cell.Width {
			cell.Text = Truncate(cell.Text, cell.Width)
			textWidth = cell.Width
		}

		if cell.Align == RightAlign {
			gridLine += strings.Repeat(" ", cell.Width-textWidth) + cell.Text
			continue
		}

		gridLine += cell.Text + strings.Repeat(" ", cell.Width-textWidth)
	}
	return gridLine

}

func Truncate(old string, n int) string {
	var (
		new       string
		newlength int
		isansi    bool
	)
	if n <= 0 {
		return new
	}
	for _, c := range old {
		if c == ansi.Marker {
			// ANSI escape sequence
			isansi = true
		} else if isansi {
			if ansi.IsTerminator(c) {
				// ANSI sequence terminated
				isansi = false
			}
		} else {
			new += string(c)
			newlength += runewidth.RuneWidth(c)
			if newlength >= n {
				return new
			}
		}
	}
	return old
}

func logfile(v ...interface{}) {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(v...)
}

func JoinLines(texts ...string) string {
	return strings.Join(
		texts,
		"\n",
	)
}
