package main

import (
	"bytes"

	"github.com/ikaito-com/lotusui/site/docspages"
	"github.com/yuin/goldmark"
	gmext "github.com/yuin/goldmark/extension"
)

func loadDocsGroups() []docspages.Group {
	changelogHTML := ""
	if md := loadChangelogMarkdown(); md != "" {
		changelogHTML = renderMarkdownString(md)
	}
	return docspages.Groups(changelogHTML, loadPerformancePage())
}

func renderMarkdownString(src string) string {
	md := goldmark.New(goldmark.WithExtensions(gmext.GFM))
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "<p>" + src + "</p>"
	}
	return buf.String()
}
