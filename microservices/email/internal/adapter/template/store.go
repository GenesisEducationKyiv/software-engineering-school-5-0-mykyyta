package template

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"email/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type templateParts struct {
	Subject *template.Template
	Plain   *template.Template
	HTML    *template.Template
}

type Store struct {
	templates map[domain.TemplateName]templateParts
}

func Load(path string) (*Store, error) {
	store := &Store{
		templates: make(map[domain.TemplateName]templateParts),
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read template dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".tmpl")
		fullPath := filepath.Join(path, entry.Name())

		tmpl, err := template.ParseFiles(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		subject := tmpl.Lookup("subject")
		plain := tmpl.Lookup("plain")
		html := tmpl.Lookup("html")

		if subject == nil || plain == nil || html == nil {
			return nil, fmt.Errorf("template %s must define subject, plain, and html blocks", name)
		}

		store.templates[domain.TemplateName(name)] = templateParts{
			Subject: subject,
			Plain:   plain,
			HTML:    html,
		}
	}

	return store, nil
}

func (s *Store) Render(ctx context.Context, templateName domain.TemplateName, data map[string]string) (subject, plain, html string, err error) {
	logger := loggerPkg.From(ctx)

	tmpl, ok := s.templates[templateName]
	if !ok {
		logger.Error("Template not found",
			"template", templateName,
			"available", len(s.templates))
		return "", "", "", fmt.Errorf("template %s not found", templateName)
	}

	logger.Debug("Rendering template", "template", templateName)

	var subjectBuf, plainBuf, htmlBuf bytes.Buffer

	if err := tmpl.Subject.Execute(&subjectBuf, data); err != nil {
		logger.Error("Template render failed",
			"template", templateName,
			"part", "subject")
		return "", "", "", fmt.Errorf("render subject: %w", err)
	}
	if err := tmpl.Plain.Execute(&plainBuf, data); err != nil {
		logger.Error("Template render failed",
			"template", templateName,
			"part", "plain")
		return "", "", "", fmt.Errorf("render plain: %w", err)
	}
	if err := tmpl.HTML.Execute(&htmlBuf, data); err != nil {
		logger.Error("Template render failed",
			"template", templateName,
			"part", "html")
		return "", "", "", fmt.Errorf("render html: %w", err)
	}

	return subjectBuf.String(), plainBuf.String(), htmlBuf.String(), nil
}
