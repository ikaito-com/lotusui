package docspages

// Groups returns the full docs IA — sidebar group order.
// changelogHTML is pre-rendered GFM for the Changelog page (empty → placeholder).
// performance is the Performance page (from bench.json); if nil, omitted from Guides.
func Groups(changelogHTML string, performance *Page) []Group {
	guides := []*Page{
		layoutPage(), responsivePage(), iconsPage(), seamlessPage(), principlesPage(),
	}
	if performance != nil {
		guides = append(guides, performance)
	}
	return []Group{
		{Title: "Get started", Pages: []*Page{
			installationPage(), quickstartPage(), registryPage(), platformsPage(), ChangelogPage(changelogHTML),
		}},
		{Title: "Theming", Pages: []*Page{
			themePage(), darkModePage(), typographyPage(),
		}},
		{Title: "Guides", Pages: guides},
		{Title: "Components", Pages: []*Page{
			accordionPage(), alertPage(), alertDialogPage(), annotatedTextPage(), avatarPage(),
			badgePage(), breadcrumbPage(), buttonPage(), buttonGroupPage(), cardPage(),
			checkboxPage(), codeBlockPage(), dialogPage(), menuPage(), examplePage(), fieldPage(), gridPage(),
			hoverCardPage(),
			inputPage(), inputOTPPage(), itemPage(), listViewPage(), paginationPage(), popoverPage(),
			progressPage(), radioGroupPage(), scrollAreaPage(), selectPage(), separatorPage(),
			simpleGridPage(), skeletonPage(), sliderPage(), spinnerPage(),
			splitPage(), stackPage(), switchPage(), tablePage(), tabsPage(),
			textareaPage(), toastPage(), togglePage(), tooltipPage(), wrapPage(),
		}},
	}
}
