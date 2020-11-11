package kunst

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// png
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/gorilla/schema"
	"github.com/ihleven/errors"

	"github.com/disintegration/imageorient"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var media string

func KunstHandler(database, medien string) http.Handler {
	media = medien
	repo, _ := NewRepo(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("asfdasdfasdf")

	})
	r.Get("/bilder", func(w http.ResponseWriter, r *http.Request) {
		bilder, _ := BilderList(repo)
		fmt.Println("bilder", bilder)
		bytes, _ := json.MarshalIndent(bilder, "", "    ")
		w.Write(bytes)

	})
	r.Post("/bilder", func(w http.ResponseWriter, r *http.Request) {
		bild, err := BilderUpdate(repo, "", r)
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		bytes, err := json.MarshalIndent(bild, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)
	})
	r.Get("/bilder/{ID}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")
		bilder, err := BilderDetail(repo, id, r)
		fmt.Println("bild =>", id, bilder, err)
		bytes, err := json.MarshalIndent(bilder, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)
	})
	r.Post("/bilder/{ID}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")
		bild, err := BilderUpdate(repo, id, r)
		fmt.Println("BilderUpdate:", id, bild, err)
		bytes, err := json.MarshalIndent(bild, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)
	})
	r.Post("/bilder/{ID}/fotos", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")
		foto, err := UploadFoto(repo, id, r)
		fmt.Println("foto", foto, err)
		bytes, err := json.MarshalIndent(foto, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)
	})
	return r
}

func BilderList(db *Repo) ([]Bild, error) {

	bilder, err := db.LoadBilder()
	if err != nil {
		return nil, err
	}

	return bilder, nil
}

var decoder = schema.NewDecoder()

func BilderDetail(db *Repo, idstr string, r *http.Request) (*Bild, error) {

	// addCorsHeader(w)
	// if r.Method == "OPTIONS" {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Println("id", idstr)
	id, err := strconv.Atoi(idstr)
	if err != nil {
		// if idstr == "neu" {
		id = 0
		fmt.Println("id", id)
		// } else {
		// 	return nil, errors.NewWithCode(400, "Invalid id")
		// }
	}

	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(0)
		if err != nil {
			return nil, err
		}
		var bild Bild
		err = decoder.Decode(&bild, r.PostForm)
		if err != nil {
			return nil, errors.Wrap(err, "Error decoding form")
		}
		if id == 0 {
			id, err = db.InsertBild(&bild)
		} else {
			err = db.SaveBild(id, &bild)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Error db ")
		}
	}
	fmt.Println("iasdf")

	bild, err := db.LoadBild(id)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bild %v", id)
	}

	return bild, nil

}

func BilderUpdate(db *Repo, idstr string, r *http.Request) (*Bild, error) {

	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}

	var bild Bild
	err = decoder.Decode(&bild, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}

	id, err := strconv.Atoi(idstr)
	if err == nil {
		err = db.SaveBild(id, &bild)
		if err != nil {
			return nil, errors.Wrap(err, "Error saving bild %v", id)
		}
	} else {
		id, err = db.InsertBild(&bild)
		if err != nil {
			return nil, errors.Wrap(err, "Error inserting bild %v", id)
		}
	}

	b, err := db.LoadBild(id)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bild %v", id)
	}

	return b, nil

}

func UploadFileBinary(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile(media, "binary-*")
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

func UploadFoto(db *Repo, idstr string, r *http.Request) (*Foto, error) {

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH,OPTIONS")
	// w.Header().Set("Access-Control-Allow-Headers", "*")

	r.ParseMultipartForm(10 << 20)

	bildID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		return nil, err
	}
	fmt.Printf("File upload for bild %v --- %s\n", bildID, idstr)

	file, header, err := r.FormFile("photo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	name, err := save(file, header, bildID, media)
	if err != nil {
		return nil, err
	}

	config, format, err := getConfig(file)
	if err != nil {
		return nil, err
	}
	fmt.Println("config:", config, format, err)

	// TRANSAKTION: cancel wenn bild nicht gespeichert werden kann
	fotoID, err := db.InsertFoto(
		bildID, 0, header.Filename, 0, strings.TrimPrefix(name, media+"/"),
		format, config.Width, config.Height, time.Now(), "Caption", "kommentar",
	)
	if err != nil {
		fmt.Println("insert error:", err)

		return nil, err
	}
	fmt.Println("fotoID:", fotoID)

	err = generateThumbnail(file, fotoID, media+"/thumbs/")
	if err != nil {
		return nil, err
	}

	err = generateThumbnail100(file, fotoID, config.Width, config.Height, media+"/thumbs/100/")
	if err != nil {
		return nil, err
	}

	foto, err := db.LoadFoto(fotoID)
	return foto, err
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
	config, format, err := imageorient.DecodeConfig(file)
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

func generateThumbnail100(file io.ReadSeeker, id, width, height int, prefix string) error {

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	// img, _, err := image.Decode(file)
	img, _, err := imageorient.Decode(file)
	if err != nil {
		return err
	}

	b := img.Bounds()
	width = b.Max.X
	height = b.Max.Y

	// fmt.Println("width = ", width, height)

	if width < height {
		height = 100 * height / width
		width = 100
	} else {
		width = 100 * width / height
		height = 100
	}
	// dst = imaging.Thumbnail(img, width, height, imaging.CatmullRom)
	dst := imaging.Resize(img, width, height, imaging.Lanczos)
	path := fmt.Sprintf("%s%d.png", prefix, id)
	fmt.Println(width, height, path)
	err = imaging.Save(dst, path)
	if err != nil {
		fmt.Println("saving error:", dst, err)

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
