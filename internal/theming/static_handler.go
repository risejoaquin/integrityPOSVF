package theming

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/solidbit/integritypos/assets"
)

func StaticHandler(customPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/static/")
		
		if customPath != "" {
			customFile := filepath.Join(customPath, "static", reqPath)
			if _, err := os.Stat(customFile); err == nil {
				http.ServeFile(w, r, customFile)
				return
			}
		}
		
		embedPath := "default/static/" + reqPath
		data, err := assets.DefaultAssets.ReadFile(embedPath)
		if err == nil {
			// Inferir el content type de forma simple (para produccion se usaria http.ServeContent)
			ctype := "text/plain"
			if strings.HasSuffix(reqPath, ".css") {
				ctype = "text/css"
			} else if strings.HasSuffix(reqPath, ".js") {
				ctype = "application/javascript"
			} else if strings.HasSuffix(reqPath, ".png") {
				ctype = "image/png"
			}
			w.Header().Set("Content-Type", ctype)
			w.Write(data)
			return
		}
		
		// Fallback simple a 404
		if err != nil && os.IsNotExist(err) || err.(*fs.PathError) != nil {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func SetupStaticHandler(mux *http.ServeMux, customPath string) {
	mux.HandleFunc("GET /static/", StaticHandler(customPath))
}
