package drive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func (f *DriveAction) FileAction(r *http.Request) error {

	fmt.Printf(" - FileAction => %s: %v\n", r.Method, f.path)

	switch r.Method {
	case http.MethodPut:
		return f.FileJsonUpdateAction(r)
	case http.MethodPost:
		fmt.Println(" => HTTP Post - Content-Type:", r.Header.Get("Content-Type"))
		return f.File.UploadContent(r)
	}
	bytes, err := f.File.GetContent()
	f.Content = string(bytes)
	fmt.Println(f.Content)
	return err
}

func (f DriveAction) FileJsonUpdateAction(r *http.Request) error {
	// name, mode, owner, group als Updatedandidaten
	var body File
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return err
	}
	//if body.Content != "" {
	//io.WriteString(a.File, body.Content)
	//}

	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	return err
	// }
	// //i.update(body)
	// fmt.Println("CONTENT:", string(body))
	return nil
}

func (f *File) GetContent() ([]byte, error) { //offset, limit int) (e error) {

	fmt.Printf(" - GetContent\n")

	var content = make([]byte, f.Size)

	bytes, err := f.Read(content)
	if err != nil {
		return nil, err
	}

	if int64(bytes) != f.Size {
		return content, errors.Errorf("read %d bytes, expected %d bytes", bytes, f.Size)
	}
	return content, nil
}

// UploadContent will directly write the http post body content as content for file
func (f *File) UploadContent(r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	fmt.Println("body:", string(body))

	// fd := a.File.OpenFile(0)
	// defer fd.Close()
	// fd.Seek(0, 0)

	bytes, err := f.Write(body)
	if err != nil {
		return err
	}
	f.Size = int64(bytes)
	fmt.Printf("%n bytes written\n", bytes)

	return nil
}
