package editor

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ApplyRegionsHTML(source, mainSelector string, allowedBlockTags, allowedInlineTags []string, regions map[string]string) (string, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("parse source html: %w", err)
	}

	root, matchedSelector := findMainRoot(doc, mainSelector)
	if root == nil {
		return "", fmt.Errorf("main selector %q not found", mainSelector)
	}
	allowedBlocks := toSet(allowedBlockTags)
	if matchedSelector != "body" {
		root = refineEditableRoot(root, allowedBlocks)
	}
	assignEditableIDs(root, allowedBlocks)

	for id, regionHTML := range regions {
		target := findNodeByEditorID(root, id)
		if target == nil {
			continue
		}
		sanitized, err := sanitizeInlineHTML(regionHTML, allowedInlineTags)
		if err != nil {
			return "", err
		}
		for child := target.FirstChild; child != nil; {
			next := child.NextSibling
			target.RemoveChild(child)
			child = next
		}
		for _, node := range sanitized {
			target.AppendChild(node)
		}
	}

	removeEditorAttrs(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("render updated html: %w", err)
	}
	return buf.String(), nil
}

func sanitizeInlineHTML(regionHTML string, allowedInlineTags []string) ([]*html.Node, error) {
	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(regionHTML), context)
	if err != nil {
		return nil, fmt.Errorf("parse region html: %w", err)
	}

	allowedInlines := toSet(allowedInlineTags)
	var out []*html.Node
	for _, node := range nodes {
		out = append(out, sanitizeInlineNode(node, allowedInlines)...)
	}
	return out, nil
}

func sanitizeInlineNode(node *html.Node, allowedInlines map[string]struct{}) []*html.Node {
	switch node.Type {
	case html.TextNode:
		return []*html.Node{{Type: html.TextNode, Data: node.Data}}
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		_, inlineAllowed := allowedInlines[tag]
		if !inlineAllowed {
			children := sanitizeInlineChildren(node, allowedInlines)
			if shouldEmitLineBreak(tag, children) {
				children = append(children, &html.Node{Type: html.ElementNode, Data: "br", DataAtom: atom.Br})
			}
			return children
		}
		cloned := &html.Node{Type: html.ElementNode, Data: tag}
		cloned.Attr = sanitizeAttrs(tag, node.Attr)
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			children := sanitizeInlineNode(child, allowedInlines)
			for _, item := range children {
				cloned.AppendChild(item)
			}
		}
		return []*html.Node{cloned}
	default:
		return nil
	}
}

func sanitizeInlineChildren(node *html.Node, allowedInlines map[string]struct{}) []*html.Node {
	var out []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		out = append(out, sanitizeInlineNode(child, allowedInlines)...)
	}
	return out
}

func shouldEmitLineBreak(tag string, children []*html.Node) bool {
	switch tag {
	case "div", "p", "section", "article", "header", "footer", "li":
	default:
		return false
	}

	for i := len(children) - 1; i >= 0; i-- {
		node := children[i]
		if node.Type == html.TextNode && strings.TrimSpace(node.Data) == "" {
			continue
		}
		return !(node.Type == html.ElementNode && strings.EqualFold(node.Data, "br"))
	}
	return false
}

func sanitizeAttrs(tag string, attrs []html.Attribute) []html.Attribute {
	out := make([]html.Attribute, 0, len(attrs))
	for _, attr := range attrs {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		if key == "" {
			continue
		}
		switch {
		case strings.HasPrefix(key, "on"):
			continue
		case key == "contenteditable", key == "spellcheck":
			continue
		case key == "data-editable", key == "data-name", key == "data-editor-id", key == "data-editor-tag", key == "data-editor-scope":
			continue
		default:
			out = append(out, html.Attribute{Namespace: attr.Namespace, Key: attr.Key, Val: attr.Val})
		}
	}
	return out
}

func assignEditableIDs(root *html.Node, allowed map[string]struct{}) {
	var seq int
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if _, ok := allowed[tag]; ok && isEditableLeaf(node, allowed) {
				seq++
				id := fmt.Sprintf("node-%04d", seq)
				setAttr(node, "data-editor-id", id)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
}

func findNodeByEditorID(root *html.Node, id string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil {
			return
		}
		if node.Type == html.ElementNode && attr(node, "data-editor-id") == id {
			found = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return found
}

func removeEditorAttrs(node *html.Node) {
	if node.Type == html.ElementNode {
		filtered := node.Attr[:0]
		for _, item := range node.Attr {
			switch strings.ToLower(item.Key) {
			case "data-editable", "data-name", "data-editor-id", "data-editor-tag", "data-editor-scope":
				continue
			default:
				filtered = append(filtered, item)
			}
		}
		node.Attr = filtered
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		removeEditorAttrs(child)
	}
}
