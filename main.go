package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

var (
	host         string
	printVersion bool
	invokeURI    = "/2015-03-31/functions/function/invocations"
	version      = "unset"
	commit       = "unset"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	flag.StringVar(&host, "host", "", "host to proxy requests to as GET requests")
	flag.BoolVar(&printVersion, "v", false, "print version")

	flag.Parse()

	if printVersion {
		fmt.Printf("%s-%s", version, commit)
		os.Exit(0)
	}

	if host == "" {
		log.Println("must provide host to proxy requests to, --host")
		os.Exit(1)
	}

	log.Println("proxying requests to", host)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path

		if path == "/" {
			path = ""
		}

		urlToCall := fmt.Sprintf("http://%s%s%s", host, invokeURI, path)

		log.Println("proxying to", urlToCall)

		req, _ := http.NewRequest("POST", urlToCall, nil)

		copyHeaders(req.Header, r.Header)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	http.ListenAndServe("localhost:3000", r)

	return nil
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
