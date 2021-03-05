package formhelpers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/ihleven/errors"
)

type File struct {
	Name        string
	ContentType string
	FileSize    int64
}

func ParseForm(r *http.Request) (*File, error) {

	const MB = 1 << 20
	if err := r.ParseMultipartForm(5 * MB); err != nil {
		return nil, errors.Wrap(err, "Couldn‘t parse multipart form in %d MB", 5)
	}

	// Limit upload size
	// r.Body = http.MaxBytesReader(w, r.Body, 5*MB) // 5 Mb

	var form File
	//
	file, multipart, _ := r.FormFile("file")
	if multipart != nil {

		form.Name = multipart.Filename
		fmt.Printf("Name: %#v\n", form.Name)

		// Create a buffer to store the header of the file in
		fileHeader := make([]byte, 512)
		// Copy the headers into the FileHeader buffer
		if _, err := file.Read(fileHeader); err != nil {
			return nil, errors.Wrap(err, "Couldn‘t read fileHeader")
		}
		form.ContentType = http.DetectContentType(fileHeader)
		fmt.Printf("MIME: %#v\n", form.ContentType)

		// set position back to start.
		if _, err := file.Seek(0, 0); err != nil {
			// return err
		}
	}

	type Sizer interface {
		Size() int64
	}
	if f, ok := file.(Sizer); ok {
		form.FileSize = f.Size()
	}
	fmt.Printf("Size: %#v\n", form.FileSize)

	return &form, nil
}

func ParseInto(r *http.Request, target interface{}, ignoreUnknownKeys bool) error {

	r.ParseForm()

	decoder := schema.NewDecoder()
	if ignoreUnknownKeys {
		decoder.IgnoreUnknownKeys(true)
	}
	err := decoder.Decode(&target, r.PostForm)
	return err
}
