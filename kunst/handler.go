package kunst

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"text/template"

	// png
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/ihleven/errors"
)

func NewHandler() *Handler {
	repo, _ := NewRepo()
	return &Handler{repo: repo}
}

type Handler struct {
	repo *Repo
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	bilder, err := h.repo.LoadBilder()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	bytes, _ := json.MarshalIndent(bilder, "", "    ")

	w.Write(bytes)
}

func Bilder(h *Handler) http.HandlerFunc {
	tmpl, err := template.ParseFiles("tmplt/index.html")
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		bilder, err := h.repo.LoadBilder()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// bytes, _ := json.MarshalIndent(bilder, "", "    ")
		// w.Write(bytes)
		err = tmpl.Execute(w, map[string]interface{}{"bilder": bilder})
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "Template error:").Error()))
		}
	}
}

func BildDetail(h *Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		i, _ := strconv.Atoi(r.URL.Path[1:])
		bilder, err := h.repo.LoadBild(i)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		bytes, err := json.MarshalIndent(bilder, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)
	}
}

func UploadFileBinary(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile("temp-images", "binary-*")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	n, err := io.Copy(file, r.Body)
	if err != nil {
		panic(err)
	}
	w.Write([]byte(fmt.Sprintf("%d bytes reveived", n)))
}

func UploadFile(h *Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")

		r.ParseMultipartForm(10 << 20)

		bildID, err := strconv.Atoi(r.Form.Get("id"))
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Printf("File upload for bild %v\n", bildID)

		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer file.Close()

		name, err := save(file, header, bildID, "temp-images")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		config, err := getConfig(file)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fotoID, err := h.repo.InsertFoto(bildID, name, "", config.Width, config.Height)
		if err != nil {
			fmt.Fprintln(w, "err:", err)
			return
		}

		err = generateThumbnail(file, fotoID, "temp-images/thumbs/")
		if err != nil {
			fmt.Fprintln(w, "err:", err)
			return
		}

		fmt.Fprintf(w, "succes uplooad %s", name)
	}
}
func save(file io.ReadSeeker, header *multipart.FileHeader, id int, path string) (string, error) {
	suffixes, err := mime.ExtensionsByType(header.Header["Content-Type"][0])
	suffix := suffixes[0]

	tempfile, err := ioutil.TempFile(path, fmt.Sprintf("%d-*%s", id, suffix))
	if err != nil {
		return "", err
	}
	defer tempfile.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	n, err := tempfile.Write(bytes)
	if err != nil {
		return "", err
	}
	if n != int(header.Size) {
		return "", errors.New("written bytes and upload file size differ: %v != %v", n, header.Size)
	}

	return tempfile.Name(), nil
}

func getConfig(file io.ReadSeeker) (*image.Config, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		errors.Wrap(err, "error in DecodeConfig")
	}
	fmt.Printf("format: %v\n", format)
	return &config, nil
}

func generateThumbnail(file io.ReadSeeker, id int, prefix string) error {

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	var dst *image.NRGBA

	dst = imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
	path := fmt.Sprintf("%s%d.png", prefix, id)
	err = imaging.Save(dst, path)
	if err != nil {
		return err
	}
	return nil
}
