package kunst

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/schema"
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"
)

type BildHandler struct {
	repo  *Repo
	drive *hidrive.Drive
}

func (b *BildHandler) Dispatch(w http.ResponseWriter, r *http.Request, id int, authuser string) error {

	var err error
	var response interface{}

	switch r.Method {
	case "GET":
		if id == 0 {
			response, err = b.List(r.URL.Query(), authuser)
		} else {
			response, err = b.Retrieve(id)
		}

	case "POST", "PUT":
		const MB = 1 << 20
		if err := r.ParseMultipartForm(5 * MB); err != nil {
			return err
		}

		// Limit upload size
		r.Body = http.MaxBytesReader(w, r.Body, 5*MB) // 5 Mb

		//
		file, multipartFileHeader, err := r.FormFile("file")
		if file != nil && multipartFileHeader != nil {

			fmt.Println("=================", err, multipartFileHeader == nil)
			// Create a buffer to store the header of the file in
			fileHeader := make([]byte, 512)

			// Copy the headers into the FileHeader buffer
			if _, err := file.Read(fileHeader); err != nil {
				return err
			}

			// set position back to start.
			if _, err := file.Seek(0, 0); err != nil {
				return err
			}
			type Sizer interface {
				Size() int64
			}
			// log.Printf("Name: %#v\n", multipartFileHeader.Filename)
			// log.Printf("Size: %#v\n", file.(Sizer).Size())
			// log.Printf("MIME: %#v\n", http.DetectContentType(fileHeader))
			response, err = b.Upload(file, r.Form, id, authuser)

			if err != nil {
				return errors.Wrap(err, "Couldn‘t load updated/uploaded bild: %v", id)
			}
		} else {

			response, err = b.parseFormSubmitBildCreateOrUpdate(r, id, authuser)
		}

	case "PATCH":
		response, err = b.Update(id, r)

	case "DELETE":
		err = b.Delete(id, authuser)
	}

	if err == nil {
		render(response, w)
	}
	return errors.Wrap(err, "")
}

func (b *BildHandler) parseFormSubmitBildCreateOrUpdate(r *http.Request, id int, authuser string) (*Bild, error) {

	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}

	var bild Bild

	err = schema.NewDecoder().Decode(&bild, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}
	if id != 0 && bild.ID != 0 && id != bild.ID {
		return nil, errors.NewWithCode(400, "id in url (%s) and bild (%d) differ", id, bild.ID)
	}

	if id == 0 {
		bildWithID, err := b.CreateBild(&bild, authuser)
		// id, err = b.repo.InsertBild(&bild)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t insert bild: %v", bild)
		}
		id = bildWithID.ID
		// err = b.repo.Update("bild", id, map[string]interface{}{"dir": fmt.Sprintf("bilder/%d", id)})
		// if err != nil {
		// 	return nil, errors.Wrap(err, "Couldn‘t update bild directory: %v", bild)
		// }
		// _, err = b.drive.Mkdir("bilder", strconv.Itoa(id), authuser)
		// if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
		// 	return nil, errors.Wrap(err, "Couldn‘t mkdir bilder/%s", id)
		// }

	} else {
		err = b.repo.SaveBild(id, &bild)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t save bild: %v", bild)
		}
	}

	bildptr, err := b.repo.LoadBild(id)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t load bild: %v", id)
	}
	return bildptr, nil
}

func (h *BildHandler) Upload(file multipart.File, form url.Values, id int, authuser string) (*Bild, error) {

	var bild *Bild
	var err error
	if id != 0 {
		bild, err = h.repo.LoadBild(id)
		if err != nil {
			err = errors.Wrap(err, "Error loading bild %v", id)
		}
	} else {
		// leeres Bild erzeugen
		bild, err = h.CreateBild(new(Bild), authuser)
		if err != nil {
			err = errors.Wrap(err, "Error creating empty bild")
		}
		id = bild.ID
	}

	defer file.Close()

	meta, err := h.drive.CreateFile(bild.Directory, file, form.Get("name"), form.Get("modtime"), authuser)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var dtorig time.Time
	if t, err := time.Parse("2006:01:02 15:04:05", meta.Image.Exif.DateTimeOriginal); err == nil {
		dtorig = t
	}
	unescapedName, _ := url.QueryUnescape(meta.Name)
	_, err = h.repo.InsertFoto(
		id, 0, unescapedName, int(meta.Size), path.Join(bild.Directory, unescapedName),
		meta.MIMEType, meta.Image.Width, meta.Image.Height, dtorig, meta.ID, "", //fmt.Sprintf("%#v", meta.Image.Exif),
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return h.repo.LoadBild(id)
}

func (b *BildHandler) CreateBild(bild *Bild, authuser string) (*Bild, error) {
	id, err := b.repo.InsertBild(bild)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t insert bild: %v", bild)
	}
	bild.ID = id

	_, err = b.drive.Mkdir("bilder", strconv.Itoa(id), authuser)
	if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
		return nil, errors.Wrap(err, "Couldn‘t mkdir bilder/%s", id)
	}

	err = b.repo.Update("bild", id, map[string]interface{}{"dir": fmt.Sprintf("bilder/%d", id)})
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t update bild directory: %v", bild)
	}
	return b.repo.LoadBild(id)
}
