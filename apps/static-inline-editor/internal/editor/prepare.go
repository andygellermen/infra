package editor

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type PreparedDocument struct {
	HTML         string
	EditableIDs  []string
	EditableTags []string
	RegionName   string
}

func PrepareDocument(source, mainSelector string, allowedBlockTags []string) (PreparedDocument, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return PreparedDocument{}, fmt.Errorf("parse html: %w", err)
	}

	root := findMainRoot(doc, mainSelector)
	if root == nil {
		return PreparedDocument{}, fmt.Errorf("main selector %q not found", mainSelector)
	}

	allowed := toSet(allowedBlockTags)
	root = refineEditableRoot(root, allowed)
	var ids []string
	var tags []string
	var seq int
	regionName := "main-content"
	setAttr(root, "data-editable", "")
	setAttr(root, "data-name", regionName)

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if _, ok := allowed[tag]; ok {
				seq++
				id := fmt.Sprintf("node-%04d", seq)
				setAttr(node, "data-editor-id", id)
				setAttr(node, "data-editor-tag", tag)
				setAttr(node, "data-editor-scope", "text")
				ids = append(ids, id)
				tags = append(tags, tag)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	removeScriptNodes(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return PreparedDocument{}, fmt.Errorf("render html: %w", err)
	}

	return PreparedDocument{
		HTML:         buf.String(),
		EditableIDs:  ids,
		EditableTags: tags,
		RegionName:   regionName,
	}, nil
}

func findMainRoot(doc *html.Node, selector string) *html.Node {
	selectors := splitSelectors(selector)
	if len(selectors) == 0 {
		selectors = []string{"main"}
	}

	for _, selector := range selectors {
		match := selectorMatcher(selector)
		if match == nil {
			continue
		}
		var found *html.Node
		var walk func(*html.Node)
		walk = func(node *html.Node) {
			if found != nil {
				return
			}
			if match(node) {
				found = node
				return
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
		}
		walk(doc)
		if found != nil {
			return found
		}
	}
	return nil
}

func removeScriptNodes(node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == html.ElementNode && strings.EqualFold(child.Data, "script") {
			node.RemoveChild(child)
			child = next
			continue
		}
		removeScriptNodes(child)
		child = next
	}
}

func refineEditableRoot(root *html.Node, allowed map[string]struct{}) *html.Node {
	if root == nil {
		return nil
	}
	if !strings.EqualFold(root.Data, "body") {
		return root
	}

	best := findBestEditableContainer(root, allowed)
	if best != nil {
		return best
	}
	return root
}

func findBestEditableContainer(root *html.Node, allowed map[string]struct{}) *html.Node {
	var best *html.Node
	bestScore := 0

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && node != root {
			score := editableContainerScore(node, allowed)
			if score > bestScore {
				best = node
				bestScore = score
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return best
}

func editableContainerScore(node *html.Node, allowed map[string]struct{}) int {
	tag := strings.ToLower(node.Data)
	switch tag {
	case "script", "style", "noscript", "head":
		return 0
	}

	score := 0
	switch tag {
	case "main":
		score += 100
	case "article":
		score += 80
	case "section":
		score += 60
	case "div":
		score += 40
	}

	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.ElementNode {
			if _, ok := allowed[strings.ToLower(current.Data)]; ok {
				score += 10
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return score
}

func splitSelectors(selector string) []string {
	raw := strings.Split(selector, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func selectorMatcher(selector string) func(*html.Node) bool {
	switch {
	case strings.HasPrefix(selector, "."):
		className := strings.TrimPrefix(selector, ".")
		return func(node *html.Node) bool {
			return hasClass(node, className)
		}
	case strings.HasPrefix(selector, "#"):
		id := strings.TrimPrefix(selector, "#")
		return func(node *html.Node) bool {
			return attr(node, "id") == id
		}
	default:
		tagName := strings.ToLower(selector)
		return func(node *html.Node) bool {
			return node.Type == html.ElementNode && strings.ToLower(node.Data) == tagName
		}
	}
}

func attr(node *html.Node, key string) string {
	for _, item := range node.Attr {
		if item.Key == key {
			return item.Val
		}
	}
	return ""
}

func hasClass(node *html.Node, className string) bool {
	if node.Type != html.ElementNode {
		return false
	}
	classes := strings.Fields(attr(node, "class"))
	for _, item := range classes {
		if item == className {
			return true
		}
	}
	return false
}

func setAttr(node *html.Node, key, value string) {
	for idx, item := range node.Attr {
		if item.Key == key {
			node.Attr[idx].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

func toSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		if trimmed := strings.ToLower(strings.TrimSpace(item)); trimmed != "" {
			out[trimmed] = struct{}{}
		}
	}
	return out
}
