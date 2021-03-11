package ui

import (
	"fmt"
	// "strings"
	"testing"
)

func TestLine(t *testing.T) {
	fmt.Println(
		Line(
			20,
			Cell{
				Align: LeftAlign,
				Text:  StyleGuessText("|一二三四五六七八九十"),
			},
			Cell{
				Align: RightAlign,
				Text:  StyleGuessText("|abcdefghij"),
			},
		))

	fmt.Println(
		Line(
			20,
			Cell{
				Align: LeftAlign,
				Text:  StyleGuessText("|abcdefghij"),
			},
			Cell{
				Align: RightAlign,
				Text:  StyleGuessText("|abc一二三四五六七八九十"),
			},
		))
	// text := "n. 阿帕奇人(Apache的复数形式 ,美洲印第安人s的一..."
	// fmt.Println(string([]rune("abbr. 美国政治和社会科学研究院(American ..."[:40])))
}
