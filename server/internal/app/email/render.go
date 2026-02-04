package email

import (
	"bytes"
	htmltpl "html/template"
	texttpl "text/template"

	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
)

type RenderedTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

func RenderTemplate(tpl *domainemail.Template, variables map[string]any) (RenderedTemplate, error) {
	subject, err := renderText(tpl.SubjectTemplate, variables)
	if err != nil {
		return RenderedTemplate{}, err
	}
	htmlBody, err := renderHTML(tpl.HTMLTemplate, variables)
	if err != nil {
		return RenderedTemplate{}, err
	}
	textBody, err := renderText(tpl.TextTemplate, variables)
	if err != nil {
		return RenderedTemplate{}, err
	}
	return RenderedTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}, nil
}

func renderText(source string, variables map[string]any) (string, error) {
	tpl, err := texttpl.New("text").Option("missingkey=error").Parse(source)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, variables); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderHTML(source string, variables map[string]any) (string, error) {
	tpl, err := htmltpl.New("html").Option("missingkey=error").Parse(source)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, variables); err != nil {
		return "", err
	}
	return buf.String(), nil
}
