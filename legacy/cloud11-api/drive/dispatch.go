package drive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ihleven/cloud11-api/auth"
	"github.com/pkg/errors"
)

func checkHTTPError(w http.ResponseWriter, err error) bool {
	fmt.Printf("checkerror")
	if err != nil {
		status := http.StatusInternalServerError
		cause := errors.Cause(err)
		if os.IsNotExist(cause) {
			status = http.StatusNotFound
		} else if os.IsExist(cause) {
			status = http.StatusInternalServerError
		} else if os.IsPermission(cause) {
			status = http.StatusForbidden
		} else if e, ok := cause.(*os.PathError); ok {
			switch e {

			case os.ErrClosed:
				status = http.StatusGone
			case os.ErrNoDeadline:
				status = http.StatusInternalServerError
			}

			//http.Error(w, fmt.Sprintf("---%v %v %v", e.Op, e.Path, e.Err.Error()), 500)
		}
		http.Error(w, cause.Error(), status)
		return true
	}
	return false
}

func DispatchRaw(wd Driver) http.HandlerFunc {

	var dispatchRaw http.HandlerFunc

	dispatchRaw = func(w http.ResponseWriter, r *http.Request) {

		cleanedPath := filepath.Clean(strings.Replace(r.URL.Path, "|", ".", 1))
		fmt.Println("serve raw:", cleanedPath)
		h, err := wd.Open(cleanedPath) //, os.O_RDONLY)
		if checkHTTPError(w, err) {
			return
		}
		authuser := auth.CurrentUser //, err := session.GetSessionUser(r, w)

		if !h.HasReadPermission(authuser) {
			http.Error(w, "Account '"+authuser.Username+"' has no read permission", 403)
			return
		}

		if h.IsDir() {
			r.URL.Path = path.Join(r.URL.Path, "index.html")
			dispatchRaw(w, r)
			return
		}

		fd, err := h.OpenFile(os.O_RDONLY, 0)
		if checkHTTPError(w, err) {
			return
		}
		defer fd.Close()

		http.ServeContent(w, r, path.Base(cleanedPath), h.ModTime(), fd)
	}
	return dispatchRaw
}

func DispatchHandler(wd Driver) http.HandlerFunc {

	//var action = Action(wd)
	//var driveAction = DriveAction{Drive: wd}

	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		action := DriveAction{
			path:    filepath.Clean(strings.Replace(r.URL.Path, "|", ".", 1)),
			Drive:   wd,
			Account: auth.CurrentUser,
		}

		file, err := wd.OpenFile(action.path, auth.CurrentUser)
		if err != nil {
			HandleDriveError(err, w)
			return
		}

		action.File = file

		if file.Type.Filetype == "D" {
			//action.FolderAction(r)
			action.Entries, err = wd.ListFiles(action.File, auth.CurrentUser)
		}
		if file.Type.Mediatype == "text" {

			err = action.FileAction(r)
		}
		if file.Type.Mediatype == "image" {

			err = action.ImageAction(r)
		}
		if err != nil {
			HandleDriveError(err, w)
		}
		// if checkHTTPError(w, err) {
		// 	return
		// }
		action.Breadcrumbs = generateParents(action.File.URL, "root")

		js, err := json.MarshalIndent(action, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func Dispatch(wd Driver) func(*http.Request) (*DriveAction, error) {

	return func(r *http.Request) (action *DriveAction, err error) {

		action = &DriveAction{
			path:    filepath.Clean(strings.Replace(r.URL.Path, "|", ".", 1)),
			Drive:   wd,
			Account: auth.CurrentUser,
		}

		action.File, err = wd.OpenFile(action.path, auth.CurrentUser)
		if err != nil {
			return nil, err
		}

		//var action func(r *http.Request, file *File) (*DriveResponse, error)

		if action.File.Mode.IsDir() {
			action.FolderAction(r)
			action.Entries, err = wd.ListFiles(action.File, auth.CurrentUser)
		}
		if action.File.Type.Mediatype == "text" {
			err = action.FileAction(r)

			// var bytecontent []byte
			// bytecontent, err = action.File.GetContent()
			// action.Content = string(bytecontent)
		}
		if action.File.Type.Mediatype == "image" {
			fmt.Println("image handler", r.Method)
			err = action.ImageAction(r)
		}
		if err != nil {
			return nil, err
		}
		action.Breadcrumbs = generateParents(action.File.URL, "root")

		return
	}
}
func DispatchHome(wd Driver) func(*http.Request) (*DriveAction, error) {

	return func(r *http.Request) (action *DriveAction, err error) {

		if auth.CurrentUser.HomeDir == "" {
			return nil, errors.New("Authemtication required")
		}

		action = &DriveAction{
			path:    filepath.Clean(strings.Replace(r.URL.Path, "|", ".", 1)),
			Drive:   wd,
			Account: auth.CurrentUser,
		}

		action.File, err = wd.OpenFile(action.path, auth.CurrentUser)
		if err != nil {
			return nil, err
		}

		//var action func(r *http.Request, file *File) (*DriveResponse, error)

		if action.File.Mode.IsDir() {
			action.FolderAction(r)
			action.Entries, err = wd.ListFiles(action.File, auth.CurrentUser)
		}
		if action.File.Type.Mediatype == "text" {
			// err = driveAction.TextFileAction(r, file)
			var bytecontent []byte
			bytecontent, err = action.File.GetContent()
			action.Content = string(bytecontent)
		}
		if action.File.Type.Mediatype == "image" {
			//err = action.ImageAction(r)
		}
		if err != nil {
			return nil, err
		}
		action.Breadcrumbs = generateParents(action.File.URL, "root")

		return
	}
}

func generateParents(url, name string) []Breadcrumb {
	var path string
	elements := strings.Split(url[1:], "/")
	list := make([]Breadcrumb, len(elements)+1)
	list[0] = Breadcrumb{Name: name, URL: "/"}
	for index, element := range elements {
		path += ("/" + element)
		list[index+1] = Breadcrumb{Name: element, URL: path}
	}
	return list
}
func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, GET, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
