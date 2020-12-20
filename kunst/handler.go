package kunst

import (
	"encoding/json"
	"fmt"
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
		response, err = b.parseFormSubmitBildCreateOrUpdate(r, id, authuser)

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
		id, err = b.repo.InsertBild(&bild)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t insert bild: %v", bild)
		}

		_, err = b.drive.Mkdir("bilder", strconv.Itoa(id), authuser)
		if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
			fmt.Println("err:", err)
			return nil, errors.Wrap(err, "Couldn‘t mkdir bilder/%s", id)
		}

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

func (b *BildHandler) List(query url.Values, authuser string) ([]Bild, error) {

	var phase string
	if p := query.Get("phase"); Schaffensphase(p).IsValid() {
		phase = p
	}
	bilder, err := b.repo.LoadBilder(phase, "")
	if err != nil {
		err = errors.Wrap(err, "error in bilder list")
	}
	return bilder, nil
}

func (b *BildHandler) Retrieve(id int) (*Bild, error) {

	bild, err := b.repo.LoadBild(id)
	if err != nil {
		err = errors.Wrap(err, "Error loading bild %v", id)
	}
	return bild, err
}

func (b *BildHandler) Update(id int, r *http.Request) (*Bild, error) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t parse form")
	}
	if foto := r.Form.Get("foto_id"); foto != "" {
		err = b.repo.Update("bild", id, map[string]interface{}{"foto_id": foto})
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t update bild %d", id)
		}
	}

	bild, err := b.repo.LoadBild(id)
	return bild, errors.Wrap(err, "Couldn‘t load bild %d", id)
}

func (b *BildHandler) Delete(id int, authuser string) error {

	err := b.repo.Delete("foto", "bild_id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete fotos of bild %d", id)
	}
	err = b.repo.Delete("bild", "id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete bild %d", id)
	}
	err = b.drive.Rmdir("bilder/"+strconv.Itoa(id), authuser)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete hidrive bild folder %d", id)
	}
	// w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *BildHandler) Upload(w http.ResponseWriter, r *http.Request, id int, authuser string) error {

	// func Upload( ) {

	r.ParseMultipartForm(10 << 20)

	name := r.Form.Get("name")
	modtime := r.Form.Get("modtime")

	file, header, err := r.FormFile("file")
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer file.Close()

	dirname := "bilder/" + strconv.Itoa(id)
	meta, err := h.drive.CreateFile(dirname, file, name, modtime, authuser)
	if err != nil {
		return errors.Wrap(err, "")
	}
	fmt.Printf("meta: %#v %v\n", meta, header.Filename)
	// meta, err = drive.GetMeta(path.Join(dir, meta.Name), username)
	// fmt.Printf("meta2: %#v %v -> %v\n", meta, header.Filename, err)
	// if err != nil {
	// 	http.Error(w, errors.Wrap(err, "error:").Error(), 500)
	// 	return
	// }

	var dtorig time.Time
	if t, err := time.Parse("2006:01:02 15:04:05", meta.Image.Exif.DateTimeOriginal); err == nil {
		dtorig = t
	}
	u, err := h.repo.InsertFoto(
		id, 0, meta.Name, int(meta.Size), path.Join(dirname, name),
		meta.MIMEType, meta.Image.Width, meta.Image.Height, dtorig, meta.ID, "", //fmt.Sprintf("%#v", meta.Image.Exif),
	)
	fmt.Println("id", u)
	if err != nil {
		return errors.Wrap(err, "")
	}
	render(meta, w)
	return nil
}

type SerieHandler struct {
	repo  *Repo
	drive *hidrive.Drive
}

func (h *SerieHandler) Dispatch(w http.ResponseWriter, r *http.Request, id int, authuser string) error {

	var err error
	var response interface{}

	switch r.Method {
	case "GET":
		if id == 0 {
			response, err = h.List(r.URL.Query(), authuser)
		} else {
			response, err = h.Retrieve(id)
		}

	case "POST", "PUT":
		response, err = h.Create(r, authuser)

	case "PATCH":
		response, err = h.Update(id, r)

	case "DELETE":
		err = h.Delete(id, authuser)
	}

	if err == nil {
		render(response, w)
	}
	return errors.Wrap(err, "")
}

func (h *SerieHandler) List(query url.Values, authuser string) ([]Serie, error) {

	return h.repo.LoadSerien()
}

func (h *SerieHandler) Create(r *http.Request, authuser string) (*Serie, error) {

	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}

	var serie Serie
	err = schema.NewDecoder().Decode(&serie, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}

	id, err := h.repo.InsertSerie(&serie)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	_, err = h.drive.Mkdir("serien", strconv.Itoa(id), authuser)
	if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
		return nil, errors.Wrap(err, "Couldn‘t mkdir serien/%s", id)
	}

	return h.repo.LoadSerie(id)
}

func (h *SerieHandler) Retrieve(id int) (*Serie, error) {
	serie, err := h.repo.LoadSerie(id)
	if err != nil {
		return nil, errors.Wrap(err, "error:")
	}

	return serie, nil
}

func (h *SerieHandler) Update(id int, r *http.Request) (*Serie, error) {

	var fieldmap map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&fieldmap)
	if err != nil {
		return nil, errors.Wrap(err, "error:")
	}
	defer r.Body.Close()

	serie, err := h.repo.UpdateSerie(id, fieldmap)
	if err != nil {
		return nil, errors.Wrap(err, "error:")
	}
	return serie, nil
}

func (h *SerieHandler) Delete(id int, authuser string) error {

	err := h.repo.Delete("foto", "serie_id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete fotos of serie %d", id)
	}
	err = h.repo.Delete("serie", "id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete serie %d", id)
	}
	err = h.drive.Rmdir("serien/"+strconv.Itoa(id), authuser)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete hidrive serien folder %d", id)
	}
	// w.WriteHeader(http.StatusNoContent)
	return nil
}
