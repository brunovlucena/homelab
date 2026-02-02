package eventing

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateRenderer handles rendering of eventing templates
type TemplateRenderer struct {
	templates map[string]*template.Template
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() (*TemplateRenderer, error) {
	renderer := &TemplateRenderer{
		templates: make(map[string]*template.Template),
	}

	templateFiles := []string{
		"broker.yaml.tmpl",
		"triggers.yaml.tmpl",
		"dlq.yaml.tmpl",
		"rbac.yaml.tmpl",
		"apisource.yaml.tmpl",
	}

	funcMap := template.FuncMap{
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 || val == false {
				return defaultVal
			}
			return val
		},
	}

	for _, file := range templateFiles {
		content, err := templatesFS.ReadFile("templates/" + file)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", file, err)
		}

		tmpl, err := template.New(file).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", file, err)
		}

		renderer.templates[file] = tmpl
	}

	return renderer, nil
}

// RenderBroker renders the broker template
func (r *TemplateRenderer) RenderBroker(data BrokerData) (string, error) {
	return r.render("broker.yaml.tmpl", data)
}

// RenderTriggers renders the triggers template
func (r *TemplateRenderer) RenderTriggers(data TriggerData) (string, error) {
	return r.render("triggers.yaml.tmpl", data)
}

// RenderDLQ renders the DLQ resources template
func (r *TemplateRenderer) RenderDLQ(data DLQData) (string, error) {
	return r.render("dlq.yaml.tmpl", data)
}

// RenderRBAC renders the RBAC template
func (r *TemplateRenderer) RenderRBAC(data RBACData) (string, error) {
	return r.render("rbac.yaml.tmpl", data)
}

// RenderApiSource renders the ApiServerSource template
func (r *TemplateRenderer) RenderApiSource(data ApiSourceData) (string, error) {
	return r.render("apisource.yaml.tmpl", data)
}

func (r *TemplateRenderer) render(templateName string, data interface{}) (string, error) {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}
