package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

const csDir = "cheatsheets"
const csTmplDir = "www"
const alpineURL = "//unpkg.com/alpinejs@3.4.2/dist/cdn.min.js"

func newCsMarkdownParser() *parser.Parser {
	extensions := parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings |
		parser.AutoHeadingIDs |
		parser.HeadingIDs |
		parser.NoEmptyLineBeforeBlock
	return parser.NewWithExtensions(extensions)
}

func csBuildToc(doc ast.Node, path string) []*tocNode {
	//logf("csBuildToc: %s\n", path)
	//ast.Print(os.Stdout, doc)

	taken := map[string]bool{}
	ensureUniqueID := func(id string) {
		panicIf(taken[id], "duplicate heading id '%s' in '%s'", id, path)
		taken[id] = true
	}

	var currHeading *ast.Heading
	var currHeadingContent string
	var allHeaders []*tocNode
	var currToc *tocNode
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		switch v := node.(type) {
		case *ast.Heading:
			if entering {
				currHeading = v
			} else {
				ensureUniqueID(currHeading.HeadingID)
				tn := &tocNode{
					heading:      currHeading,
					Content:      currHeadingContent,
					ID:           currHeading.HeadingID,
					HeadingLevel: currHeading.Level,
				}
				allHeaders = append(allHeaders, tn)
				currToc = tn
				currHeading = nil
				currHeadingContent = ""
				//headingLevel := currHeading.Level
			}
		case *ast.Text:
			// the only child of ast.Heading is ast.Text (I think)
			if currHeading != nil && entering {
				currHeadingContent = string(v.Literal)
			} else {
				if entering && currToc != nil {
					currToc.SiblingsCount++
				}
			}
		default:
			if entering && currToc != nil {
				currToc.SiblingsCount++
			}
		}
		return ast.GoToNext
	})

	if false {
		for _, tn := range allHeaders {
			logf(ctx(), "h%d #%s %s %d siblings\n", tn.HeadingLevel, tn.heading.HeadingID, tn.Content, tn.SiblingsCount)
		}
	}
	cloneNode := func(n *tocNode) *tocNode {
		// clone but without children
		return &tocNode{
			heading:       n.heading,
			Content:       n.Content,
			HeadingLevel:  n.HeadingLevel,
			ID:            n.ID,
			SiblingsCount: n.SiblingsCount,
			Class:         n.Class,
		}
	}

	buildToc := func() []*tocNode {
		first := cloneNode(allHeaders[0])
		toc := []*tocNode{first}
		stack := []*tocNode{first}
		for _, node := range allHeaders[1:] {
			node = cloneNode(node)
			stackLastIdx := len(stack) - 1
			curr := stack[stackLastIdx]
			currLevel := curr.HeadingLevel
			nodeLevel := node.HeadingLevel
			if nodeLevel > currLevel {
				// this is a child
				// TODO: should synthesize if we skip more than 1 level?
				panicIf(nodeLevel-currLevel > 1, "skipping more than 1 level in %s, '%s'", path, node.Content)
				curr.Children = append(curr.Children, node)
				stack = append(stack, node)
				curr = node
			} else if nodeLevel == currLevel {
				// this is a sibling, make current and attach to
				stack[stackLastIdx] = node
				if stackLastIdx > 0 {
					parent := stack[stackLastIdx-1]
					parent.Children = append(parent.Children, node)
				} else {
					toc = append(toc, node)
				}
			} else {
				// nodeLevel < currLevel
				for stackLastIdx > 0 {
					if stackLastIdx == 1 {
						toc = append(toc, node)
						stack = []*tocNode{node}
						stackLastIdx = 0
					} else {
						stack = stack[:stackLastIdx]
						stackLastIdx--
						curr = stack[stackLastIdx]
						if curr.HeadingLevel == nodeLevel {
							stack[stackLastIdx] = node
							parent := stack[stackLastIdx-1]
							parent.Children = append(parent.Children, node)
							stackLastIdx = 0
						}
					}
				}
			}
		}
		// remove intro if at the top level
		for i, node := range toc {
			if node.ID == "intro" && len(node.Children) == 0 {
				toc = append(toc[:i], toc[i+1:]...)
				return toc
			}
		}
		return toc
	}
	toc := buildToc()
	if false {
		printToc(toc, 0)
	}

	// set alternating colors clas
	for i, node := range toc {
		cls := "bgcol1"
		if i%2 == 1 {
			cls = "bgcol2"
		}
		node.Class = cls
		for _, c := range node.Children {
			c.Class = cls
		}
	}
	return toc
}

func printToc(nodes []*tocNode, indent int) {
	indentStr := func(indent int) string {
		return "............................"[:indent]
	}
	hdrStr := func(level int) string {
		return "#################"[:level]
	}

	for _, n := range nodes {
		s := indentStr(indent)
		hdr := hdrStr(n.HeadingLevel)
		logf(ctx(), "%s%s %s\n", s, hdr, n.Content)
		printToc(n.Children, indent+1)
	}
}

var reg *regexp.Regexp

func init() {
	reg = regexp.MustCompile(`{:.*}`)
}

func cleanupMarkdown(md []byte) []byte {
	s := string(md)
	// TODO: implement support of this in markdown parser
	// remove lines like: {: data-line="1"}
	s = reg.ReplaceAllString(s, "")
	s = strings.Replace(s, "{% raw %}", "", -1)
	s = strings.Replace(s, "{% endraw %}", "", -1)
	prev := s
	for prev != s {
		prev = s
		s = strings.Replace(s, "\n\n", "\n", -1)
	}
	return []byte(s)
}

type cheatSheet struct {
	fileNameBase string // unique name from file name, without extension
	mdFileName   string // path relative to . directory
	mdPath       string
	htmlFullPath string
	// TODO: rename htmlFileName
	PathHTML   string // path relative to . directory
	mdWithMeta []byte
	md         []byte
	meta       map[string]string
	Title      string
	inMain     bool // if true shown in /index.html, if false only in /all.html
}

func processCheatSheet(cs *cheatSheet) {
	//logf("processCheatSheet: '%s'\n", cs.mdPath)
	cs.mdWithMeta = readFileMust(cs.mdPath)
	md := normalizeNewlinesInPlace(cs.mdWithMeta)
	lines := strings.Split(string(md), "\n")
	// skip empty lines at the beginning
	for len(lines[0]) == 0 {
		lines = lines[1:]
	}
	if lines[0] != "---" {
		// no metadata
		cs.md = []byte(strings.Join(lines, "\n"))
		return
	}
	metaLines := []string{}
	lines = lines[1:]
	for lines[0] != "---" {
		metaLines = append(metaLines, lines[0])
		lines = lines[1:]
	}
	lines = lines[1:]
	cs.md = []byte(strings.Join(lines, "\n"))
	//logf("meta for '%s':\n%s\n", cs.mdPath, strings.Join(metaLines, "\n"))
	lastName := ""
	for _, line := range metaLines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 1 {
			s := strings.TrimSpace(parts[0])
			s = strings.Trim(s, `"`)
			v := cs.meta[lastName]
			if len(v) > 0 {
				v = v + "\n"
			}
			v += s
			cs.meta[lastName] = v
		} else {
			name := parts[0]
			name = strings.TrimSpace(name)
			name = strings.ToLower(name)
			s := strings.TrimSpace(parts[1])
			s = strings.Trim(s, `"`)
			s = strings.TrimLeft(s, "|")
			cs.meta[name] = s
			lastName = name
		}
	}
	cs.Title = cs.meta["title"]
	if cs.Title == "" {
		cs.Title = cs.fileNameBase
	}
}

type tocNode struct {
	heading *ast.Heading // not set if synthesized

	Content      string
	HeadingLevel int
	TocLevel     int
	ID           string
	Class        string

	SiblingsCount int

	Children []*tocNode // level of child is > our level

	tocHTML      []byte
	tocHTMLBlock *ast.HTMLBlock
	seen         bool
}

func genHeadingTocHTML(node *tocNode, level int) {
	nChildren := len(node.Children)
	buildToc := func() {
		shouldBuild := ((level >= 2) || (node.SiblingsCount == 0))
		if nChildren == 0 || !shouldBuild {
			return
		}

		gen := func(active *tocNode) []byte {
			s := `<div class="toc-mini">`
			for i, c := range node.Children {
				if c == active {
					s += fmt.Sprintf(`<b>%s</b>`, c.Content)
				} else {
					s += fmt.Sprintf(`<a href="#%s">%s</a>`, c.ID, c.Content)
				}
				if i < nChildren-1 {
					s += `<span class="tmb">&bull;</span>`
				}
			}
			s += `</div>`
			return []byte(s)
		}

		//logf("genTocHTML: generating for %s '%s'\n", node.ID, s)
		node.tocHTML = gen(nil)
		node.tocHTMLBlock = &ast.HTMLBlock{Leaf: ast.Leaf{Literal: node.tocHTML}}

		// TODO: generates too much
		// for _, c := range node.Children {
		// 	c.tocHTML = gen(c)
		// 	c.tocHTMLBlock = &ast.HTMLBlock{Leaf: ast.Leaf{Literal: c.tocHTML}}
		// }

	}
	buildToc()

	for _, c := range node.Children {
		genHeadingTocHTML(c, level+1)
	}
}

func findTocNodeForHeading(toc []*tocNode, h *ast.Heading) *tocNode {
	for _, n := range toc {
		if n.heading == h {
			return n
		}
		// deapth first search
		if c := findTocNodeForHeading(n.Children, h); c != nil {
			return c
		}
	}
	return nil
}

// for 2nd+ level headings we need to create a toc-mini pointing to its children
func insertAutoToc(doc ast.Node, toc []*tocNode) {

	for _, n := range toc {
		genHeadingTocHTML(n, 1)
	}

	// doc is ast.Document, all ast.Heading are direct childre
	// we fish out the ast.Heading and insert tocHTMLBlock after ast.Heading
	onceMore := true
	for onceMore {
		onceMore = false
		a := doc.GetChildren()
		for i, n := range a {
			hn, ok := n.(*ast.Heading)
			if !ok {
				continue
			}
			tn := findTocNodeForHeading(toc, hn)
			if tn == nil || tn.seen {
				continue
			}
			tn.seen = true
			if tn.tocHTMLBlock != nil {
				//logf("inserting toc for heading %s\n", hn.HeadingID)
				insertAstNodeChild(doc, tn.tocHTMLBlock, i+1)
				tn.tocHTMLBlock = nil
				// re-do from beginning if modified
				onceMore = true
			}
		}
	}
}

// insertAstNodeChild appends child to children of parent
// It panics if either node is nil.
func insertAstNodeChild(parent ast.Node, child ast.Node, i int) {
	//ast.RemoveFromTree(child)
	child.SetParent(parent)
	a := parent.GetChildren()
	if i >= len(a) {
		a = append(a, child)
	} else {
		a = append(a[:i], append([]ast.Node{child}, a[i:]...)...)
	}
	parent.SetChildren(a)
}

func buildFlatToc(toc []*tocNode, tocLevel int) []*tocNode {
	res := []*tocNode{}
	for _, n := range toc {
		n.TocLevel = tocLevel
		res = append(res, n)
		sub := buildFlatToc(n.Children, tocLevel+1)
		res = append(res, sub...)
	}
	return res
}

func genCheatsheetHTML(cs *cheatSheet) []byte {
	logf(ctx(), "csGenHTML: for '%s'\n", cs.mdPath)
	md := cleanupMarkdown(cs.md)

	parser := newCsMarkdownParser()

	doc := markdown.Parse(md, parser)
	toc := csBuildToc(doc, cs.mdPath)
	tocFlat := buildFlatToc(toc, 0)

	// [[text, text.toLowerCase(), id, tocLevel], ...]
	searchIndex := [][]interface{}{}
	for _, toc := range tocFlat {
		s := toc.Content
		v := []interface{}{s, strings.ToLower(s), toc.ID, toc.TocLevel}
		searchIndex = append(searchIndex, v)
	}

	insertAutoToc(doc, toc)
	//ast.Print(os.Stdout, doc)
	renderer := newMarkdownHTMLRenderer("")
	mdHTML := string(markdown.Render(doc, renderer))

	tpl := string(readFileMust(filepath.Join(csTmplDir, "cheatsheet.tmpl.html")))

	// on windows mdFileName is a windows-style path so change to unix/url style
	mdFileName := strings.Replace(cs.mdFileName, `\`, "/", -1)

	searchIndexJSON, err := json.Marshal(searchIndex)
	must(err)

	ctx := map[string]interface{}{
		"toc": toc,
		//"tocflat":    tocFlat,
		"title":             cs.Title,
		"mdFileName":        mdFileName,
		"content":           mdHTML,
		"searchIndexStatic": string(searchIndexJSON),
		"alpineURL":         alpineURL,
	}

	return []byte(raymond.MustRender(tpl, ctx))
}

func genIndexHTML(cheatsheets []*cheatSheet) string {
	// sort by title
	sort.Slice(cheatsheets, func(i, j int) bool {
		t1 := strings.ToLower(cheatsheets[i].Title)
		t2 := strings.ToLower(cheatsheets[j].Title)
		return t1 < t2
	})

	byCat := map[string][]*cheatSheet{}
	for _, cs := range cheatsheets {
		cat := cs.meta["category"]
		if cat == "" {
			continue
		}
		byCat[cat] = append(byCat[cat], cs)
	}

	// build toc for categories
	categories := []string{}
	for cat := range byCat {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	cats := []map[string]interface{}{}
	for _, category := range categories {
		v := map[string]interface{}{}
		v["category"] = category
		catMetas := byCat[category]
		v["cheatsheets"] = catMetas
		cats = append(cats, v)
	}

	tpl := string(readFileMust(filepath.Join(csTmplDir, "index.tmpl.html")))
	ctx := map[string]interface{}{
		"cheatsheets":      cheatsheets,
		"CheatsheetsCount": len(cheatsheets),
		"categories":       cats,
		"alpineURL":        alpineURL,
	}
	s := raymond.MustRender(tpl, ctx)
	return s
}

func readCheatSheets() []*cheatSheet {
	logvf(ctx(), "readCheatSheets\n")
	cheatsheets := []*cheatSheet{}

	readFromDir := func() {
		filepath.WalkDir(csDir, func(path string, f fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if f.IsDir() {
				return nil
			}
			name := f.Name()
			if filepath.Ext(name) != ".md" {
				return nil
			}
			baseName := strings.Split(name, ".")[0]
			cs := &cheatSheet{
				fileNameBase: baseName,
				mdPath:       path,
				mdFileName:   path, // TODO: something else?
				meta:         map[string]string{},
				inMain:       strings.Contains(path, "good"),
			}

			//logf("%s\n", cs.mdPath)
			cheatsheets = append(cheatsheets, cs)
			return nil
		})
	}

	readFromDir()

	{
		// uniquify names
		taken := map[string]bool{}
		for _, cs := range cheatsheets {
			name := cs.fileNameBase
			n := 0
			for taken[name] {
				n++
				name = fmt.Sprintf("%s%d", cs.fileNameBase, n)
			}
			taken[name] = true
			cs.fileNameBase = name
		}
	}

	for _, cs := range cheatsheets {
		cs.PathHTML = cs.fileNameBase // + ".html"
		cs.htmlFullPath = filepath.Join(csDir, cs.PathHTML)
	}

	nThreads := runtime.NumCPU()
	//nThreads := 1
	sem := make(chan bool, nThreads)
	var wg sync.WaitGroup
	for _, cs := range cheatsheets {
		wg.Add(1)
		sem <- true
		go func(cs *cheatSheet) {
			processCheatSheet(cs)
			//logf("Processed %s, html size: %d\n", cs.mdPath, len(cs.html))
			wg.Done()
			<-sem
		}(cs)
	}
	wg.Wait()

	logf(ctx(), "%d cheatsheets\n", len(cheatsheets))
	return cheatsheets
}
