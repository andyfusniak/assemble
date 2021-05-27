package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andyfusniak/assemble/assembler"
	"github.com/andyfusniak/assemble/manifest"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	compilePages := flag.Bool("c", false, "compile pages and write to outputDir")
	watch := flag.Bool("w", false, "watch templateDir for changes and recompile")
	verbose := flag.Bool("verbose", false, "enable verbose mode")
	debug := flag.Bool("debug", false, "enable debug mode")
	port := flag.String("port", "9000", "")
	httpLog := flag.Bool("httplog", false, "enabled HTTP logging")
	flag.Parse()

	if !*compilePages && !*watch {
		fmt.Fprint(os.Stdout, "Assemble is the main command used to build your Assemble site.\n\n")
		fmt.Fprint(os.Stdout, "usage: assemble [flags]\n\n")
		fmt.Fprint(os.Stdout, "flags:\n")
		fmt.Fprintf(os.Stdout, "  -c        compile pages and write to outputDir\n")
		fmt.Fprintf(os.Stdout, "  -w        watch templateDir for changes and recompile\n")
		fmt.Fprintf(os.Stdout, "  -httplog  watch templateDir for changes and recompile\n")
		fmt.Fprintf(os.Stdout, "  -port     specify a specific port for the built-in HTTP server")
		fmt.Fprintf(os.Stdout, "  -verbose  enable verbose mode\n")
		fmt.Fprintf(os.Stdout, "  -debug    enable debug mode\n")
		os.Exit(0)
	}

	if *compilePages && *watch {
		fmt.Fprintf(os.Stderr, "Choose either -c (compile pages) or -w (watch and recompile) only\n")
		os.Exit(1)
	}

	assemble, err := manifest.LoadAssembleFile("assemble.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load assemble.json: %+v", err)
		os.Exit(1)
	}

	// check for missing templates
	_, err = assemble.AllTemplates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get list of all templates")
		os.Exit(1)
	}
	// fmt.Println(allTemplates)

	opts := &assembler.Opts{
		VerboseMode: *verbose,
		DebugMode:   *debug,
	}
	assembly, err := assembler.NewAssembly(assemble.AssetsDir, assemble.TemplateDir, assemble.OutputDir, opts)
	if err != nil {
		log.Fatal(err)
	}

	for name, entry := range assemble.Targets {
		assembly.NewTarget(name, entry.Path, entry.Templates)
	}

	if *compilePages {
		if err := assembly.WriteTargets(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write targets: %+v\n", err)
			os.Exit(1)
		}
	}

	if *watch {
		if err := assembly.Compile(); err != nil {
			log.Fatal(err)
		}

		// static assets
		r := chi.NewRouter()
		if *httpLog {
			r.Use(middleware.Logger)
		}
		http.Handle("/static/", http.FileServer(http.Dir("./dist")))
		filesDir := http.Dir("./public")
		FileServer(r, "/static", filesDir)

		// dynamic routes
		routeMaps := assembly.Routes()
		for path, name := range routeMaps {
			r.Get(path, buildHandler(assemble.OutputDir, name))
		}

		err := assembler.Watch("/home/andy/projects/namwa/icnc/icnc-ui-template/templates")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to watch for changes: %+v\n", err)
			os.Exit(1)
		}

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%s", *port),
			Handler: r,
		}
		fmt.Printf("HTTP Server running on http://localhost:%s\n", *port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func buildHandler(outputDir, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(outputDir, name))
	}
}
