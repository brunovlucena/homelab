package build

import (
	"embed"
)

// Embed all template files from the templates directory
//
//go:embed templates/runtimes/nodejs/runtime.js.tmpl
//go:embed templates/runtimes/python/runtime.py.tmpl
//go:embed templates/runtimes/go/runtime.go.tmpl
//go:embed templates/dockerfiles/nodejs.Dockerfile.tmpl
//go:embed templates/dockerfiles/python.Dockerfile.tmpl
//go:embed templates/dockerfiles/go.Dockerfile.tmpl
//go:embed templates/scripts/*.tmpl
var templatesFS embed.FS

// GetRuntimeTemplate returns the runtime template for the given language
func GetRuntimeTemplate(language string) (string, error) {
	var path string
	switch language {
	case "nodejs", "node", "javascript", "js":
		path = "templates/runtimes/nodejs/runtime.js.tmpl"
	case "python", "py":
		path = "templates/runtimes/python/runtime.py.tmpl"
	case "go", "golang":
		path = "templates/runtimes/go/runtime.go.tmpl"
	default:
		path = "templates/runtimes/nodejs/runtime.js.tmpl"
	}

	data, err := templatesFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetDockerfileTemplate returns the Dockerfile template for the given language
func GetDockerfileTemplate(language string) (string, error) {
	var path string
	switch language {
	case "nodejs", "node", "javascript", "js":
		path = "templates/dockerfiles/nodejs.Dockerfile.tmpl"
	case "python", "py":
		path = "templates/dockerfiles/python.Dockerfile.tmpl"
	case "go", "golang":
		path = "templates/dockerfiles/go.Dockerfile.tmpl"
	default:
		path = "templates/dockerfiles/nodejs.Dockerfile.tmpl"
	}

	data, err := templatesFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetScriptTemplate returns a script template by name
func GetScriptTemplate(name string) (string, error) {
	path := "templates/scripts/" + name + ".tmpl"
	data, err := templatesFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
