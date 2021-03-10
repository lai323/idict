package ui

import (
	"fmt"
	"testing"
)

func TestLine(t *testing.T) {
	fmt.Println(
		Line(
			100,
			Cell{
				Align: LeftAlign,
				Text:  "|" + StyleGuessText("阿波罗应用计划(Apollo Applica...") + "|",
			},
			Cell{
				Align: RightAlign,
				Text:  "|" + StyleGuessText("阿波罗应用计划(Apollo Applica...") + "|",
			},
		))
}
