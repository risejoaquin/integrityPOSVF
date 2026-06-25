package theming

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/solidbit/integritypos/assets"
)

func LoadTemplates(customPath string) (*template.Template, error) {
	tmpl := template.New("")

	// Cargar primero de los embebidos
	err := fs.WalkDir(assets.DefaultAssets, "default/templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".html" {
			data, err := assets.DefaultAssets.ReadFile(path)
			if err != nil {
				return err
			}
			name := filepath.Base(path)
			_, err = tmpl.New(name).Parse(string(data))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sobrescribir con custom templates si existen
	if customPath != "" {
		customTmplPath := filepath.Join(customPath, "templates")
		if _, err := os.Stat(customTmplPath); err == nil {
			err = filepath.Walk(customTmplPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && filepath.Ext(path) == ".html" {
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					name := filepath.Base(path)
					_, err = tmpl.New(name).Parse(string(data))
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return tmpl, nil
}
