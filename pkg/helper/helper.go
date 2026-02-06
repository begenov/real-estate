package helper

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func GeneratePlaceholders(count, startIndex int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
	}
	return strings.Join(placeholders, ", ")
}

const (
	minEmailLen = 3
	maxEmailLen = 255
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func IsEmailValid(email string) bool {
	if len(email) < minEmailLen || len(email) > maxEmailLen {
		return false
	}

	return emailRegex.MatchString(email)
}

func StripHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr
	}

	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return strings.TrimSpace(buf.String())
}
