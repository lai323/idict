package utils

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func FmtErrorf(msg string, err error) error {
	return fmt.Errorf("%s: %s", msg, err.Error())
}

// 不取子节点的内容
// InnerText returns the text between the start and end tags of the object.
func InnerTextWithOutChild(n *html.Node) string {
	var output func(*bytes.Buffer, *html.Node)
	output = func(buf *bytes.Buffer, n *html.Node) {
		switch n.Type {
		case html.TextNode:
			buf.WriteString(n.Data)
			return
		case html.CommentNode:
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
	}

	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}

func SpaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
