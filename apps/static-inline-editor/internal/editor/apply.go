package editor

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ApplyRegionHTML(source, mainSelector string, allowedBlockTags, allowedInlineTags []string, regionHTML string) (string, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("parse source html: %w", err)
	}

	root := findMainRoot(doc, mainSelector)
	if root == nil {
		return "", fmt.Errorf("main selector %q not found", mainSelector)
	}

	sanitized, err := sanitizeRegionHTML(regionHTML, allowedBlockTags, allowedInlineTags)
	if err != nil {
		return "", err
	}

	for child := root.FirstChild; child != nil; {
		next := child.NextSibling
		root.RemoveChild(child)
		child = next
	}

	for _, node := range sanitized {
		root.AppendChild(node)
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("render updated html: %w", err)
	}
	return buf.String(), nil
}

func sanitizeRegionHTML(regionHTML string, allowedBlockTags, allowedInlineTags []string) ([]*html.Node, error) {
	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(regionHTML), context)
	if err != nil {
		return nil, fmt.Errorf("parse region html: %w", err)
	}

	allowedBlocks := toSet(allowedBlockTags)
	allowedInlines := toSet(allowedInlineTags)
	var out []*html.Node
	for _, node := range nodes {
		for _, sanitized := range sanitizeNode(node, allowedBlocks, allowedInlines, true) {
			if sanitized.Type == html.TextNode && strings.TrimSpace(sanitized.Data) == "" {
				out = append(out, sanitized)
				continue
			}
			if sanitized.Type != html.ElementNode {
				continue
			}
			if _, ok := allowedBlocks[strings.ToLower(sanitized.Data)]; !ok {
				continue
			}
			out = append(out, sanitized)
		}
	}
	return out, nil
}

func sanitizeNode(node *html.Node, allowedBlocks, allowedInlines map[string]struct{}, topLevel bool) []*html.Node {
	switch node.Type {
	case html.TextNode:
		return []*html.Node{{Type: html.TextNode, Data: node.Data}}
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		_, blockAllowed := allowedBlocks[tag]
		_, inlineAllowed := allowedInlines[tag]
		if !blockAllowed && !inlineAllowed {
			return sanitizeChildren(node, allowedBlocks, allowedInlines, topLevel)
		}
		if topLevel && !blockAllowed {
			return nil
		}

		cloned := &html.Node{Type: html.ElementNode, Data: tag}
		cloned.Attr = sanitizeAttrs(tag, node.Attr)
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			children := sanitizeNode(child, allowedBlocks, allowedInlines, false)
			for _, item := range children {
				cloned.AppendChild(item)
			}
		}
		return []*html.Node{cloned}
	default:
		return nil
	}
}

func sanitizeChildren(node *html.Node, allowedBlocks, allowedInlines map[string]struct{}, topLevel bool) []*html.Node {
	var out []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		out = append(out, sanitizeNode(child, allowedBlocks, allowedInlines, topLevel)...)
	}
	return out
}

func sanitizeAttrs(tag string, attrs []html.Attribute) []html.Attribute {
	if tag != "a" {
		return nil
	}

	out := make([]html.Attribute, 0, len(attrs))
	for _, attr := range attrs {
		switch strings.ToLower(attr.Key) {
		case "href", "title", "target", "rel":
			out = append(out, html.Attribute{Key: strings.ToLower(attr.Key), Val: attr.Val})
		}
	}
	return out
}
