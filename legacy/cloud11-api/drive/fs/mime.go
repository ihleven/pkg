package fs

import (
	"mime"
	"path"
	"strings"

	"github.com/h2non/filetype/types"
	"github.com/ihleven/cloud11-api/drive"
)

func init() {

	mime.AddExtensionType(".py", "text/python")
	mime.AddExtensionType(".go", "text/golang")
	mime.AddExtensionType(".json", "text/json")
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".ts", "text/typescript")
	mime.AddExtensionType(".dia", "text/diary")
	mime.AddExtensionType(".md", "text/markdown")
}

var dir = drive.Type{
	Filetype:  "D",
	Mediatype: "dir",
	Subtype:   "",
	MIME:      "",
	Charset:   "",
}

var file = drive.Type{
	Filetype:  "F",
	Mediatype: "file",
	Subtype:   "",
	MIME:      "file",
	Charset:   "",
}

func (fh handle) GuessMIME() drive.Type {

	if fh.IsDir() {
		return dir
	}

	ext := path.Ext(fh.Name())

	if mimestr := mime.TypeByExtension(ext); mimestr != "" {

		mime := types.NewMIME(mimestr)
		splitmimestr := strings.Split(mime.Subtype, "; charset=")
		var charset string
		if len(splitmimestr) > 1 {
			charset = splitmimestr[1]
		}
		//fmt.Println(mime, splitmimestr[0])
		return drive.Type{
			Filetype:  "F",
			Mediatype: mime.Type,
			Subtype:   splitmimestr[0],
			MIME:      mime.Type + "/" + splitmimestr[0],
			Charset:   charset,
		}
		//

	}

	// if m.Value == "" {
	// 	// m, _ = f.h2nonMatchMIME261()
	// }
	// if strings.HasSuffix(m.Subtype, "charset=utf-8") {
	// 	m.Subtype = m.Subtype[:len(m.Subtype)-15]
	// 	m.Value = m.Value[:len(m.Value)-15]
	// }
	return file
}
func GetMIMEByExtension(filename string) *drive.Type {
	ext := path.Ext(filename)
	if mimestr := mime.TypeByExtension(ext); mimestr != "" {

		mime := types.NewMIME(mimestr)
		splitmimestr := strings.Split(mime.Subtype, "; charset=")
		var charset string
		if len(splitmimestr) > 1 {
			charset = splitmimestr[1]
		}
		//fmt.Println(mime, splitmimestr[0])
		return &drive.Type{
			Filetype:  "F",
			Mediatype: mime.Type,
			Subtype:   splitmimestr[0],
			MIME:      mime.Type + "/" + splitmimestr[0],
			Charset:   charset,
		}
	}
	return nil
}
