package drive

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

func (f *DriveAction) FolderAction(r *http.Request) error {

	switch r.Method {
	case http.MethodDelete:
		// delete folder recursive
	case http.MethodPut:
		// ?
	case http.MethodPost:
		fmt.Println(" => HTTP Post - Content-Type:", r.Header.Get("Content-Type"))
		err := r.ParseMultipartForm(2000000)
		if r.MultipartForm != nil && err == nil {
			err = f.UploadMultipartFormFiles(r.MultipartForm)
			if err != nil {
				return err
			}
		} else {

			fmt.Println("   => Error ParseMultipartForm:", err, r.Form)
			//return err
			if r.Form != nil {
				f.FormAction(r.Form)
			}

		}
	}
	return nil
}

func (f *DriveAction) UploadMultipartFormFiles(formdata *multipart.Form) error {
	if !f.Authorization.W {
		return fmt.Errorf("Missing write permissions for %s", f.URL)
	}
	// if formdata.File["file"] == nil {
	// 	return nil
	// }

	fmt.Println(" => Multiple file upload:", formdata.File["file"])
	//return a.FolderUploadMultipleFiles(f, formdata)

	for _, header := range formdata.File["file"] {

		file, err := header.Open()
		defer file.Close()
		if err != nil {
			return errors.Wrapf(err, "Could not open form file %v", header)
		}

		h, err := f.Drive.Create(f.URL + "/" + header.Filename, URLPath)
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

func (f *DriveAction) FormAction(form url.Values) error {

	//fmt.Println(" => FormValueMap:", form)
	//for key, value := range formdata.Value["foo"] {
	//fmt.Println(" * value:", form["foo"])

	fmt.Println(" => FormValueMap:", form)
	//for key, value := range formdata.Value["foo"] {
	fmt.Println(" * value:", form["foo"])
	for _, name := range form["folders"] {
		path := path.Join(f.URL, name)
		fmt.Println(" * folder:", f.File, path)
		fh, err := f.Drive.Mkdir(path)
		fmt.Println("%v %v", fh, err)
	}
	return nil
}
