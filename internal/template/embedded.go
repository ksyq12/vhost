package template

import "embed"

//go:embed nginx/*.tmpl
var nginxTemplates embed.FS
