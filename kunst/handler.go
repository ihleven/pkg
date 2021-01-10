package kunst

import (
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

func (b *BildHandler) List(query url.Values, authuser string) ([]Bild, error) {

	where := make(map[string]interface{})
	serienbilder := false

	if p := query.Get("phase"); Schaffensphase(p).IsValid() {
		where["phase"] = p
	}
	fmt.Println(strconv.Atoi(query.Get("teile")))
	if i, err := strconv.Atoi(query.Get("teile")); err == nil {
		where["teile"] = i
	}
	if s := query.Get("serie"); s == "true" {
		serienbilder = true
	}

	bilder, err := b.repo.LoadBilder(where, serienbilder, query.Get("sort"))
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

	bild, _ := b.repo.LoadBild(id)
	err := b.repo.Delete("foto", "bild_id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete fotos of bild %d", id)
	}
	err = b.repo.Delete("bild", "id", id)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t delete bild %d", id)
	}
	if bild.Serie != nil && bild.IndexFoto != nil {

		err = b.drive.DeleteFile(bild.IndexFoto.Path, authuser)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t delete foto %d", bild.IndexFoto.Path)
		}
	} else {
		err = b.drive.Rmdir("bilder/"+strconv.Itoa(id), authuser)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t delete hidrive bild folder %d", id)
		}
	}
	// w.WriteHeader(http.StatusNoContent)
	return nil
}

type SerieHandler struct {
	repo  *Repo
	drive *hidrive.Drive
}

func (h *SerieHandler) Dispatch(w http.ResponseWriter, r *http.Request, id int, tail, authuser string) error {

	var err error
	var response interface{}

	if id == 0 {
		if r.Method == "GET" {
			response, err = h.List(r.URL.Query(), authuser)
		}
		if r.Method == "POST" {
			response, err = h.Create(r, authuser)
		}
	} else if tail == "/bilder" {

		response, err = h.AddBild(id, r, authuser)
	} else {

		switch r.Method {
		case "GET":
			response, err = h.Retrieve(id)

		case "POST", "PUT":
			response, err = h.Update(id, r)

		case "PATCH":
			response, err = h.Update(id, r)

		case "DELETE":
			err = h.Delete(id, authuser)
		}
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

	var form Serie
	err = schema.NewDecoder().Decode(&form, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}

	id, err := h.repo.InsertSerie(&form)
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
	bilder, err := h.repo.LoadBilder(map[string]interface{}{"serie_id": serie.ID}, true, "")
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t load bilder for serie %d", id)
	}
	serie.Bilder = bilder

	return serie, nil
}

func (h *SerieHandler) Update(id int, r *http.Request) (*Serie, error) {

	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}
	var form Serie
	err = schema.NewDecoder().Decode(&form, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}

	err = h.repo.SaveSerie(id, &form)
	if err != nil {
		return nil, errors.Wrap(err, "error:")
	}
	return h.repo.LoadSerie(id)
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

func (h *SerieHandler) AddBild(id int, r *http.Request, authuser string) (*Bild, error) {

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, err
	}

	name := r.Form.Get("name")
	modtime := r.Form.Get("modtime")

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer file.Close()

	// jetzt haben wir das hochgeladene Bild in der Hand =>

	// Bild in DB anlegen, dazu Daten der Serie übernehmen
	serie, err := h.repo.LoadSerie(id)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	bildID, err := h.repo.InsertBild(&Bild{
		SerieID: &serie.ID,
		Technik: serie.Technik, Bildträger: serie.Träger,
		Höhe: serie.Höhe, Breite: serie.Breite, Tiefe: serie.Tiefe,
		Jahr: serie.Jahr, Schaffensphase: serie.Schaffensphase,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t insert bild")
	}
	h.repo.Update("bild", bildID, map[string]interface{}{"dir": fmt.Sprintf("serien/%d/%d", serie.ID, bildID)})
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t update bild directory: %v", bildID)
	}

	// Verzeichnis für das Bild anlegen
	dirname := "serien/" + strconv.Itoa(serie.ID)
	_, err = h.drive.Mkdir(dirname, strconv.Itoa(bildID), authuser)
	if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
		return nil, errors.Wrap(err, "Couldn‘t mkdir %s/%d", dirname, bildID)
	}

	// hochgeladene Bilddatei im neuen Verz. ablegen
	dirname = fmt.Sprintf("%s/%d", dirname, bildID)
	meta, err := h.drive.CreateFile(dirname, file, name, modtime, authuser)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t create hidrive file %s/%s", dirname, name)
	}

	var dtorig time.Time
	if t, err := time.Parse("2006:01:02 15:04:05", meta.Image.Exif.DateTimeOriginal); err == nil {
		dtorig = t
	}
	unescapedName, _ := url.QueryUnescape(meta.Name)
	fotoid, err := h.repo.InsertFoto(
		bildID, 0, unescapedName, int(meta.Size), path.Join(dirname, unescapedName),
		meta.MIMEType, meta.Image.Width, meta.Image.Height, dtorig, meta.ID, "", //fmt.Sprintf("%#v", meta.Image.Exif),
	)
	fmt.Printf("foto: %v\n", fotoid)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	bild, err := h.repo.LoadBild(bildID)
	fmt.Printf("bild: %v\n", bild)
	return bild, err
}
