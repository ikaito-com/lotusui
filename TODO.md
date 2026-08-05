# lotusui — open work

## Done recently
- Build-time composition (SelectItem/Grouped, Radio/Tab/Toggle Item*, AccordionItemOf, FieldGroup, CardHeader/Footer, ButtonGroupButton)
- SelectOption.Icon + Content (multiline plan rows)
- BreadcrumbNav (shadcn responsive ellipsis collapse)
- Docs demos: **iframe per Preview** (`gallery/#slug/N`) — no shared canvas / hover host
- Docs examples: shadcn-style **Preview | Code** tabs per example (Code = section Snippet)
- **Item** family (shadcn Item + ItemGroup/Media/Content/…)
- **DurationScale** / `WithDuration` — theme motion ladder; hover eases on Fast
- Performance page driven by `site/bench.json` + `lotusui bench-doc`

## Still open

### Docs as lotusui
- [x] **`site/docspages`**: shared page model
- [x] **`site/live`**: shared addressable demos
- [x] **`site/docsapp`**: Gio docs app
- [x] **Default site deploy** is docsapp WASM (`make -C site build` → `dist/`)
- [x] Retired HTML/CSS generator (`site/gen` deleted)
- [ ] Thin-wrap `site/gallery` onto `site/live` (drop duplicated demos)

### Parity / audit
- [ ] Full shadcn prop + example audit (behavior gaps, not just section lists)
- [ ] PARITY ⏳ leftovers (Tooltip keyboard/disabled, Skeleton card/form/table, Field choice-card, …)

### Missing families (new components)
- [ ] Empty, Combobox, Command
- [ ] Sheet/Drawer, Context Menu, Menubar, Navigation Menu, Calendar
- [ ] InputGroup as explicit family (beyond Input Start/End)
