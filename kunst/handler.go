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
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	// png
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/gorilla/schema"
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

var decoder = schema.NewDecoder()

func BildDetail(h *Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		addCorsHeader(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// w.Header().Set("Access-Control-Allow-Origin", "*")

		id, err := strconv.Atoi(r.URL.Path[1:])
		if err != nil {
			if r.URL.Path[1:] == "neu" {
				id = 0
			} else {
				http.Error(w, err.Error(), 400)
				return
			}
		}

		if r.Method == http.MethodPost {
			var bild Bild
			err := r.ParseMultipartForm(0)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			err = decoder.Decode(&bild, r.PostForm)
			if err != nil {
				fmt.Println("ERROR", err.Error())
			}
			if id == 0 {
				id, err = h.repo.InsertBild(&bild)
			} else {
				err = h.repo.SaveBild(id, &bild)
			}
			if err != nil {
				fmt.Println("ERROR", err.Error())
			}
		}

		bilder, err := h.repo.LoadBild(id)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

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

		config, format, err := getConfig(file)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// TRANSAKTION: cancel wenn bild nicht gespeichert werden kann
		fotoID, err := h.repo.InsertFoto(bildID, strings.TrimPrefix(name, "temp-images/"), "", config.Width, config.Height, format)
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
	var f *os.File
	var err error
	nconflict := 0
	ts := time.Now().Format("060102-150405")
	for i := 0; i < 10000; i++ {
		name := fmt.Sprintf("%s/%d-%s-%s", path, id, ts, header.Filename)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				fmt.Println("nconflict:", nconflict)
			}
			continue
		}
		defer f.Close()
		break
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	n, err := f.Write(bytes)
	if err != nil {
		return "", err
	}
	if n != int(header.Size) {
		return "", errors.New("written bytes and upload file size differ: %v != %v", n, header.Size)
	}

	return f.Name(), nil
}

func saveTemp(file io.ReadSeeker, header *multipart.FileHeader, id int, path string) (string, error) {
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

func getConfig(file io.ReadSeeker) (*image.Config, string, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return nil, "", err
	}
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		errors.Wrap(err, "error in DecodeConfig")
	}
	fmt.Printf("format: %v\n", format)
	return &config, format, nil
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

func addCorsHeader(res http.ResponseWriter) {
	headers := res.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
}
