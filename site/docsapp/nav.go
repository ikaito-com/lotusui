package main

import "github.com/ikaito-com/lotusui/site/docspages"

// docsGroups is filled in newDocsUI from docspages.
var docsGroups []docspages.Group

func pageTitle(slug string) string {
	if slug == "" {
		return "Home"
	}
	p, _, _, _ := docspages.Lookup(docsGroups, slug)
	if p != nil {
		return p.Title
	}
	return slug
}
