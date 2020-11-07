package main

import (
	"net/http"

	"github.com/ihleven/pkg/httpsrvr"
	"github.com/ihleven/pkg/kunst"
)

func main() {

	fs := http.FileServer(http.Dir("./temp-images"))
	srv := httpsrvr.NewServer("", 8000, false, false, nil)
	handler := kunst.NewHandler()
	srv.Register("/api/bilder", handler)
	srv.Register("/api/upload", kunst.UploadFile(handler))
	// srv.Register("/bilder", kunst.Bilder(handler)) // template
	srv.Register("/api/bild", kunst.BildDetail(handler))
	srv.Register("/api/media", http.StripPrefix("/", fs))

	srv.ListenAndServe()

}
