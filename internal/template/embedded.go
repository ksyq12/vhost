package template

import (
	"embed"
	"fmt"
)

//go:embed nginx/*.tmpl
var nginxTemplates embed.FS

//go:embed apache/*.tmpl
var apacheTemplates embed.FS

// getTemplateFS returns the embed.FS for the given driver
func getTemplateFS(driverName string) (embed.FS, error) {
	switch driverName {
	case "nginx":
		return nginxTemplates, nil
	case "apache":
		return apacheTemplates, nil
	default:
		return embed.FS{}, fmt.Errorf("unknown driver: %s", driverName)
	}
}
