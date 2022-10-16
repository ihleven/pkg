package hidrive

import (
	"encoding/json"
	"io"
	"net/url"
	"path"
	"sync"

	"github.com/ihleven/pkg/errors"
)

// GetMeta concurrently retrieves file and folder meta data
func (d *Drive) GetMeta(drivepath string, token *Token) (*Meta, error) {

	fullpath := d.fullpath(drivepath, token.Alias)

	var wg sync.WaitGroup
	var dir *Meta
	var direrr error

	wg.Add(1)

	go func() {
		defer wg.Done()
		dir, direrr = d.client.GetDir(fullpath, "", "", 0, 0, "", "", token.AccessToken)
	}()

	params := url.Values{
		"path":   {fullpath},
		"fields": {metafields + "," + imagefields},
	}
	body, err := d.client.Request("GET", "/meta", params, nil, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t execute hidrive meta request")
	}
	defer body.Close()

	meta, err := d.processMetaResponse(body)
	// meta, err := d.client.GetMeta(fullpath, "", "", token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	if meta.Filetype == "dir" {
		wg.Wait()
		if direrr != nil {
			return nil, errors.Wrap(direrr, "")
		}
		for i, m := range dir.Members {
			dir.Members[i].NameURLEncoded = m.Name()
			// dir.Members[i].Path = d.drivepath(m.Path, token.Alias)
		}
		meta.Members = dir.Members
	}

	return meta, nil
}

// Meta gibt ein Meta-Objekt für den übergeben Drive-Pfad zurück
// TODO: Berechtigungen prüfen
func (d *Drive) Meta(drivepath string, token *Token) (*Meta, error) {

	fullpath := d.fullpath(drivepath, token.Alias)

	meta, err := d.client.GetMeta(fullpath, "", "", token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	meta.Path = d.drivepath(meta.Path, token.Alias)
	if meta.Path == "" {
		meta.Path = "/"
		meta.NameURLEncoded = ""
	}

	if meta.Filetype == "dir" {
		dir, err := d.client.GetDir(fullpath, "", "", 0, 0, memberfields, "", token.AccessToken)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		for i, m := range dir.Members {
			dir.Members[i].NameURLEncoded = m.Name()
			// dir.Members[i].Path = d.drivepath(m.Path, token.Alias)
		}
		meta.Members = dir.Members
	}
	// else if strings.HasPrefix(meta.MIMEType, "text") || meta.MIMEType == "application/json" || meta.MIMEType == "application/xml" {
	// 	body, err := d.client.GetFile(fullpath, "", token.AccessToken)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "Couldn't decode response body")
	// 	}

	// 	bytes, _ := io.ReadAll(body)
	// 	fmt.Println("body:", string(bytes))
	// 	meta.Content = string(bytes)

	// }

	return meta, nil
}

func (d *Drive) Mkdir(drivepath string, token *Token) (*Meta, error) {

	fullpath := d.fullpath(drivepath, token.Alias)

	dir, err := d.client.PostDir(fullpath, "", "autoname", 0, 0, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}

	return dir, nil
}

func (d *Drive) Listdir(drivepath string, token *Token) ([]Meta, error) {

	params := url.Values{
		"path":    {d.fullpath(drivepath, token.Alias)},
		"members": {"all"},
		"fields":  {memberfields},
	}

	body, err := d.client.Request("GET", "/dir", params, nil, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer body.Close()

	var dir Meta
	err = json.NewDecoder(body).Decode(&dir)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	for i, m := range dir.Members {
		dir.Members[i].NameURLEncoded = m.Name()
		dir.Members[i].Path = path.Join(drivepath, m.Name()) //d.drivepath(m.Path, token.Alias)
	}
	return dir.Members, nil
}

func (d *Drive) Rmdir(drivepath string, token *Token) error {

	params := url.Values{
		"path":      {d.fullpath(drivepath, token.Alias)},
		"recursive": {"true"},
	}

	_, err := d.client.Request("DELETE", "/dir", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}

	return nil
}

func (d *Drive) Rm(drivepath string, token *Token) error {

	params := url.Values{
		"path": {d.fullpath(drivepath, token.Alias)},
	}

	_, err := d.client.Request("DELETE", "/file", params, nil, token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Error in delete request")
	}

	return nil
}

func (d *Drive) CreateFile(drivepath string, body io.Reader, name string, modtime string, token *Token) (*Meta, error) {
	params := url.Values{
		"dir":      {d.fullpath(drivepath, token.Alias)},
		"name":     {name},
		"on_exist": {"autoname"},
	}
	if modtime != "" {
		params.Set("mtime", modtime)
	}
	response, err := d.client.Request("POST", "/file", params, body, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in post request")
	}
	defer response.Close()

	return d.processMetaResponse(response)
}

func (d *Drive) Save(drivepath string, body io.Reader, token *Token) (*Meta, error) {

	dir, file := path.Split(d.fullpath(drivepath, token.Alias))
	// dir = strings.TrimSuffix(dir, "/")
	meta, err := d.client.PutFile(body, dir, file, 0, 0, token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Error in hidrive PUT file request")
	}

	return meta, nil
}
