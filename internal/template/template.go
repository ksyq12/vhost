package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/ksyq12/vhost/internal/config"
)

// TemplateData contains data for rendering templates
type TemplateData struct {
	Domain     string
	Aliases    []string
	Root       string
	ProxyPass  string
	PHPVersion string
	SSL        bool
	SSLCert    string
	SSLKey     string
}

// Render renders a template for the given vhost and driver
func Render(driverName string, vhost *config.VHost) (string, error) {
	tmplPath := fmt.Sprintf("%s/%s.tmpl", driverName, vhost.Type)

	// Get template filesystem for the driver
	fs, err := getTemplateFS(driverName)
	if err != nil {
		return "", err
	}

	// Read template content
	content, err := fs.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("template not found: %s/%s", driverName, vhost.Type)
	}

	// Create template with custom functions
	funcMap := template.FuncMap{
		"replace": strings.ReplaceAll,
	}

	tmpl, err := template.New(vhost.Type).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		Domain:     vhost.Domain,
		Root:       vhost.Root,
		ProxyPass:  vhost.ProxyPass,
		PHPVersion: vhost.PHPVersion,
		SSL:        vhost.SSL,
		SSLCert:    vhost.SSLCert,
		SSLKey:     vhost.SSLKey,
	}

	// Set default PHP version if not specified
	if data.PHPVersion == "" {
		data.PHPVersion = "8.2"
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// Available returns all available template types for a driver
func Available(driverName string) []string {
	return config.ValidTypes()
}
