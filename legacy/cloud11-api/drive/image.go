package drive

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func (a *DriveAction) ImageAction(r *http.Request) error {

	var handle = a.File.Handle

	fmt.Printf(" - ImageAction => %s: %v\n", r.Method, a.path)

	switch r.Method {
	case http.MethodPost:
		err := a.ImagePostAction(r)
		if err != nil {
			return errors.Wrap(err, "Error in ImagePostAction")
		}
	}

	if !a.File.Authorization.R {
		return errors.Errorf("Missing read permission for %v (User %s)", a.File.Name, a.Account.Username)
	}

	image, err := handle.ReadImage()
	if err != nil {
		return errors.Wrap(err, "ReadImage")
	}
	a.Image = image

	// siblings, err := handle.Siblings()
	// if err != nil {
	// 	return errors.Wrap(err, "siblings error")
	// }

	return err
}

func (a *DriveAction) ImagePostAction(r *http.Request) error {

	if !a.File.Authorization.W {
		return errors.Errorf("Missing Write permission for %v (User %s)", a.File.Name, a.Account.Username)
	}
	image := &Image{}
	// if err != nil {
	// 	return errors.Wrap(err, "NewImage")
	// }
	a.Image = image
	if err := r.ParseMultipartForm(2000000); err != nil {
		return errors.Wrap(err, "parse form")
	}

	image.Title = r.FormValue("title")
	image.Caption = r.FormValue("caption")
	image.Cutline = r.FormValue("cutline")
	
	if err := a.WriteImageMeta(); err != nil {

		return errors.Wrap(err, "WriteImageMeta")
	}

	// http.Redirect(w, r, v.File.Path, http.StatusFound)
	return nil
}

func metaFilename(path string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return fmt.Sprintf("%s.txt", base)
}

func (a *DriveAction) WriteImageMeta() error {
	

	metaHandle, err := a.Drive.Open(metaFilename(a.path))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not Found") || os.IsNotExist(errors.Cause(err)) {
			metaHandle, err = a.Drive.Create(metaFilename(a.path), DrivePath)
		}
		if err != nil {
			return errors.Wrapf(err, "Could not open MetaFile %v", metaFilename(a.path))
		}
	}
	

	//fd, err := os.OpenFile(a.Image.MetaFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	// if i.metaFile == nil {
	// 	filename := i.getMetaFilename()
	// 	file, err := CreateFile(i.File.Storage(), filename, usr)
	// 	if err != nil {
	// 		return errors.Errorf("Could not create meta file: %s", filename)
	// 	}
	// 	i.metaFile = file
	// }

	// if perms = metaFile.GetPermissions(a.Account); !perms.W {
	// 	return errors.Errorf("Missing write permission for %s", i.metaFile.Name)
	// }
	// fd := i.metaFile.Descriptor(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
	// defer fd.Close()
	// fmt.Println("writemeta", fd.Name())

	tmpl, err := template.New("txt").Parse("{{.Title}}\n=====\n{{.Caption}}\n-----\n{{.Cutline}}\n------\n")
	if err != nil {
		return errors.Wrap(err, "Could not create meta file template")
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, a.Image)
	if err != nil {
		return errors.Wrap(err, "Could not execute template")
	}

	_, err = metaHandle.Write(tpl.Bytes())
	return err
}

// func (i *Image) update(requestBody []byte) {

// 	var m map[string]interface{}
// 	_ = json.Unmarshal(requestBody, &m)

// 	if title, ok := m["Title"]; ok {
// 		i.Title = title.(string)
// 		fmt.Printf(" - update Title => '%s'\n", title)
// 	}
// 	if caption, ok := m["Caption"]; ok {
// 		i.Caption = caption.(string)
// 		fmt.Printf(" - update Caption => '%s'\n", caption)
// 	}
// 	if cutline, ok := m["Cutline"]; ok {
// 		i.Cutline = cutline.(string)
// 		fmt.Printf(" - update Cutline => '%s'\n", cutline)
// 	}

// 	i.WriteMeta(nil)
// }
