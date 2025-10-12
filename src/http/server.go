package http

import (
	"embed"
	"fmt"
	"io/fs"
	stdhttp "net/http"
	"path"
	"strings"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
)

//go:embed ui/dist/*
var uiDist embed.FS

func StartServer(cfg *config.Config) (*stdhttp.Server, error) {
	if cfg.WebPort == 0 {
		log.Infof("Web server disabled (port 0)")
		return nil, nil
	}
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/ws/logs", wsHandler)
	dist, err := fs.Sub(uiDist, "ui/dist")
	if err == nil {
		mux.Handle("/", spa(dist))
	} else {
		mux.HandleFunc("/", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><head><title>B4</title></head><body>No UI build found</body></html>`))
		})
	}
	addr := fmt.Sprintf(":%d", cfg.WebPort)
	log.Infof("Starting web server on %s", addr)
	srv := &stdhttp.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Errorf("Web server error: %v", err)
		}
	}()
	return srv, nil
}

func spa(fsys fs.FS) stdhttp.Handler {
	fileServer := stdhttp.FileServer(stdhttp.FS(fsys))
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		upath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if upath == "" {
			upath = "index.html"
		}
		f, err := fsys.Open(upath)
		if err == nil {
			if info, e := f.Stat(); e == nil && !info.IsDir() {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			_ = f.Close()
		}
		data, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			w.WriteHeader(stdhttp.StatusInternalServerError)
			_, _ = w.Write([]byte("index.html not found"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write(data)
	})
}
