package drive

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
)

// Action takes HTTP requests (URLs and their methods)
// and uses that input to interact with the domain,
// after which it passes the domain's output to one and only one responder.
func Action(wd Driver) func(*http.Request, *File) error {

	return func(r *http.Request, file *File) error {

		switch r.Method {
		case http.MethodGet:
		case http.MethodDelete:
		case http.MethodPut:
		case http.MethodPost:
			fmt.Println(" => HTTP Post - Content-Type:", r.Header.Get("Content-Type"))
		}
		return nil
	}
}

func (a DriveAction) PostAction(r *http.Request, file *File) error {
	fmt.Println(" => PostAction - Content-Type:", r.Header.Get("Content-Type"))

	if file.Mode.IsDir() {
		err := r.ParseMultipartForm(2000000)
		if r.MultipartForm != nil && err == nil {
			return a.MultipartFormAction(r.MultipartForm, file)
		} else {

			fmt.Println("   => Error ParseMultipartForm:", err, r.Form)
			//return err
			if r.Form != nil {
				a.FormAction(r.Form)
			}

		}
	} else {
		_ = file.UploadContent(r)

	}
	return nil
}

func (a DriveAction) MultipartFormAction(formdata *multipart.Form, f *File) error {

	if !f.Authorization.W {
		return fmt.Errorf("Missing write permissions for %s", f.URL)
	}
	if formdata.File["files"] != nil {
		fmt.Println(" => Multiple file upload:", formdata.File["files"])
		return a.FolderUploadMultipleFiles(f, formdata)
	}

	return nil
}

func (a DriveAction) FolderUploadMultipleFiles(folder *File, formdata *multipart.Form) error {

	for _, header := range formdata.File["files"] {

		file, err := header.Open()
		defer file.Close()
		if err != nil {
			return errors.Wrapf(err, "Could not open form file %v", header)
		}

		h, err := a.Drive.Create(folder.URL+"/"+header.Filename, URLPath)
		if err != nil {
			return errors.Wrapf(err, "Could not upload to folder '%v'. Unable to create the file for writing. Check your write access privilege", header.Filename)
		}

		_, err = io.Copy(h, file)
		if err != nil {
			return errors.Wrapf(err, "Unable to copy formfile")
		}
		fmt.Println(" * uploaded", h)
	}
	return nil
}

// func (a *DirActionResponder) PutAction(r *http.Request, w http.ResponseWriter) error {

// 	file := a.File
// 	fmt.Printf("PutAction => Directory \"%s/\"\n", file.Name)

// 	//if !file.Permissions.Write {
// 	//	return errors.Errorf("no write permissions")
// 	//}

// 	var options struct {
// 		CreateThumbnails bool
// 	}
// 	err := json.NewDecoder(r.Body).Decode(&options)
// 	if err != nil {
// 		return errors.Wrap(err, "Error decoding put request body")
// 	}

// 	if options.CreateThumbnails {

// 		err := drive.MakeThumbs(file.Handle)
// 		if err != nil {
// 			return errors.Wrap(err, "Error making thumbnails")
// 		}
// 	}
// 	return nil
// }
