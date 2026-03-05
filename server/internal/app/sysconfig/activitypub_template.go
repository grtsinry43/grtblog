package sysconfig

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

const (
	defaultActivityPubPublishTemplate = `<p><strong>{{ .Title }}</strong></p>{{ if .Summary }}<p>{{ .Summary }}</p>{{ end }}{{ if .URL }}<p><a href="{{ .URL }}" rel="nofollow noopener noreferrer">阅读全文</a></p>{{ end }}`
	activityPubPublishTemplateKey     = "activitypub.publishTemplate"
)

var activityPubPublishTemplateAllowedFields = map[string]struct{}{
	"Title":       {},
	"Summary":     {},
	"URL":         {},
	"ContentType": {},
}

func validateActivityPubPublishTemplate(raw string) error {
	tpl := strings.TrimSpace(raw)
	if tpl == "" {
		return nil
	}
	parsed, err := template.New("activitypub.publishTemplate").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return err
	}
	for _, tree := range parsed.Templates() {
		if tree == nil || tree.Tree == nil || tree.Tree.Root == nil {
			continue
		}
		if err := validateTemplateNodeFields(tree.Tree.Root); err != nil {
			return err
		}
	}
	return nil
}

func validateTemplateNodeFields(node parse.Node) error {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *parse.ListNode:
		for _, child := range n.Nodes {
			if err := validateTemplateNodeFields(child); err != nil {
				return err
			}
		}
	case *parse.ActionNode:
		if err := validateTemplatePipeFields(n.Pipe); err != nil {
			return err
		}
	case *parse.IfNode:
		if err := validateTemplatePipeFields(n.Pipe); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.List); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.ElseList); err != nil {
			return err
		}
	case *parse.RangeNode:
		if err := validateTemplatePipeFields(n.Pipe); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.List); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.ElseList); err != nil {
			return err
		}
	case *parse.WithNode:
		if err := validateTemplatePipeFields(n.Pipe); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.List); err != nil {
			return err
		}
		if err := validateTemplateNodeFields(n.ElseList); err != nil {
			return err
		}
	case *parse.TemplateNode:
		return nil
	}
	return nil
}

func validateTemplatePipeFields(pipe *parse.PipeNode) error {
	if pipe == nil {
		return nil
	}
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			switch node := arg.(type) {
			case *parse.FieldNode:
				if len(node.Ident) == 0 {
					continue
				}
				name := strings.TrimSpace(node.Ident[0])
				if _, ok := activityPubPublishTemplateAllowedFields[name]; !ok {
					return fmt.Errorf("unsupported template variable: %s", name)
				}
			case *parse.PipeNode:
				if err := validateTemplatePipeFields(node); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
