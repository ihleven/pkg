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
	srv.Register("/", handler)
	srv.Register("/upload", kunst.UploadFile(handler))
	srv.Register("/bilder", kunst.Bilder(handler))
	srv.Register("/bild", kunst.BildDetail(handler))
	srv.Register("/media", http.StripPrefix("/", fs))

	srv.ListenAndServe()

}
