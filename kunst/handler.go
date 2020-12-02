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
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"

	"github.com/disintegration/imageorient"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var media string

func KunstHandler(database, medien string, hdclient *hidrive.HiDriveClient) http.Handler {
	media = medien
	repo, _ := NewRepo(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Liste der Ausstellungen
	r.Get("/ausstellungen", func(w http.ResponseWriter, r *http.Request) {

		ausstellungen, err := repo.LoadAusstellungen()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Println("err:", err, ausstellungen)
		bytes, err := json.MarshalIndent(ausstellungen, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(bytes)
	})

	// Neue Ausstellung anlegen
	r.Post("/ausstellungen", func(w http.ResponseWriter, r *http.Request) {
		ausstellung, err := parseFormSubmitAusstellung(r, 0)

		id, err := repo.InsertAusstellung(ausstellung)
		if err != nil {
			http.Error(w, errors.Wrap(err, "Couldn‘t insert ausstellung: %v", ausstellung).Error(), errors.Code(err))
			return
		}
		dir, err := hdclient.Mkdir("ausstellungen", strconv.Itoa(id))
		fmt.Println("dir created:", dir, errors.Code(err))
		if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
			http.Error(w, err.Error(), 500)
			return
		}
		ausstellung, err = repo.LoadAusstellung(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		bytes, err := json.MarshalIndent(ausstellung, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(bytes)
	})

	// Daten zu einer bestimmten Ausstellung
	r.Get("/ausstellungen/{ID}", func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, "parse error id", 500)
			return
		}

		ausstellung, err := repo.LoadAusstellung(id)
		if err != nil {
			http.Error(w, errors.Wrap(err, "Error loading ausstellung %v", id).Error(), errors.Code(err))
			return
		}

		fmt.Println("ausstellung =>", id, ausstellung, err)
		bytes, err := json.MarshalIndent(ausstellung, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "marshaling error").Error()))
			return
		}
		w.Write(bytes)
	})

	// Daten zu einer bestimmten Ausstellung ändern
	r.Put("/ausstellungen/{ID}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id in url: %d", chi.URLParam(r, "ID")), 400)
		}

		ausstellung, err := parseFormSubmitAusstellung(r, id)

		err = repo.SaveAusstellung(ausstellung)
		if err != nil {
			http.Error(w, fmt.Sprintf("Couldn‘t save ausstellung: %v", ausstellung), 500)
		}
		ausstellung, err = repo.LoadAusstellung(id)
		bytes, err := json.MarshalIndent(ausstellung, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		w.Write(bytes)
	})

	// Dateien zu einer bestimmten Ausstellung ins hidrive hochladen
	r.Post("/ausstellungen/{ID}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id in url: %d", chi.URLParam(r, "ID")), 400)
		}

		r.ParseMultipartForm(10 << 20)

		modtime := r.Form.Get("modtime")
		name := r.Form.Get("name")

		file, header, err := r.FormFile("file")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		defer file.Close()
		fmt.Printf("header: %s, name: %s, modtime: %v\n", header.Filename, name, modtime)
		body, err := hdclient.UploadFile(fmt.Sprintf("ausstellungen/%d", id), file, name, modtime)
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		// err = repo.SaveAusstellung(ausstellung)
		// if err != nil {
		// 	http.Error(w, fmt.Sprintf("Couldn‘t save ausstellung: %v", ausstellung), 500)
		// }
		// ausstellung, err = repo.LoadAusstellung(id)
		bytes, err := json.MarshalIndent(body, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		w.Write(bytes)
	})

	r.Get("/ausstellungen/{ID}/documents", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")
		fmt.Println("ducument:", id)
		dir, err := hdclient.GetDir("/users/matt.ihle/wolfgang-ihle/ausstellungen/"+id, nil)
		fmt.Println("ducument:", errors.Code(err), dir, err)
		if err != nil {
			if errors.Code(err) == 404 {
				w.Write([]byte("[]"))
				return
			}
			http.Error(w, errors.Wrap(err, "error:").Error(), errors.Code(err))
			return
		}
		bytes, _ := json.MarshalIndent(dir.Members, "", "    ")
		w.Write(bytes)
	})

	r.Get("/serien", func(w http.ResponseWriter, r *http.Request) {
		serien, _ := SerienList(repo)
		fmt.Println("serien", serien)
		bytes, _ := json.MarshalIndent(serien, "", "    ")
		w.Write(bytes)

	})
	r.Post("/serien", func(w http.ResponseWriter, r *http.Request) {
		serie, err := parseFormSubmitSerie(r)
		if err != nil || serie == nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		fmt.Println("geparste serie:", serie)

		id, err := repo.InsertSerie(serie)
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		serie, err = repo.LoadSerie(id)
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}

		bytes, err := json.MarshalIndent(serie, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
		}
		w.Write(bytes)

	})
	r.Get("/serien/{ID}", func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}

		serie, err := repo.LoadSerie(id)
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}

		bytes, err := json.MarshalIndent(serie, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		w.Write(bytes)
	})

	// Bilder list
	r.Get("/bilder", func(w http.ResponseWriter, r *http.Request) {
		var phase string
		if p := r.URL.Query().Get("phase"); Schaffensphase(p).IsValid() {
			phase = p
		}
		bilder, err := repo.LoadBilder(phase, "")
		if err != nil {
			http.Error(w, errors.Wrap(err, "error in bilder list").Error(), errors.Code(err))
			return
		}

		bytes, _ := json.MarshalIndent(bilder, "", "    ")
		w.Write(bytes)
	})
	// r.Post("/bilder", func(w http.ResponseWriter, r *http.Request) {
	// 	bild, err := BilderUpdate(repo, "", r)
	// 	if err != nil {
	// 		w.Write([]byte(errors.Wrap(err, "error:").Error()))
	// 	}
	// 	bytes, err := json.MarshalIndent(bild, "", "    ")
	// 	if err != nil {
	// 		w.Write([]byte(errors.Wrap(err, "error:").Error()))
	// 	}
	// 	w.Write(bytes)
	// })

	// Neues Bild anlegen
	r.Post("/bilder", func(w http.ResponseWriter, r *http.Request) {
		bild, err := parseFormSubmitBild(r, 0)

		id, err := repo.InsertBild(bild)
		if err != nil {
			http.Error(w, errors.Wrap(err, "Couldn‘t insert bild: %v", bild).Error(), errors.Code(err))
			return
		}
		dir, err := hdclient.Mkdir("bilder", strconv.Itoa(id))
		fmt.Println("dir created:", dir, errors.Code(err))
		if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
			http.Error(w, err.Error(), 500)
			return
		}
		bild, err = repo.LoadBild(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		bytes, err := json.MarshalIndent(bild, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(bytes)
	})

	r.Get("/bilder/{ID}", func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		bild, err := repo.LoadBild(id)
		if err != nil {
			http.Error(w, errors.Wrap(err, "Error loading bild %v", id).Error(), 400)
			return

		}

		bytes, err := json.MarshalIndent(bild, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		w.Write(bytes)
	})
	// r.Post("/bilder/{ID}", func(w http.ResponseWriter, r *http.Request) {
	// 	id := chi.URLParam(r, "ID")
	// 	bild, err := BilderUpdate(repo, id, r)
	// 	fmt.Println("BilderUpdate:", id, bild, err)
	// 	bytes, err := json.MarshalIndent(bild, "", "    ")
	// 	if err != nil {
	// 		w.Write([]byte(errors.Wrap(err, "error:").Error()))
	// 	}
	// 	w.Write(bytes)
	// })

	// Daten zu einem bestimmten Bild ändern
	r.Put("/bilder/{ID}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id in url: %d", chi.URLParam(r, "ID")), 400)
			return
		}

		bild, err := parseFormSubmitBild(r, id)
		if err != nil {
			http.Error(w, errors.Wrap(err, "Couldn‘t parse form").Error(), 400)
			return
		}

		err = repo.SaveBild(id, bild)
		if err != nil {
			http.Error(w, fmt.Sprintf("Couldn‘t save bild: %v", bild), 500)
		}
		bild, err = repo.LoadBild(id)
		bytes, err := json.MarshalIndent(bild, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
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

	// Dateien zu einer bestimmten Ausstellung ins hidrive hochladen
	r.Post("/bilder/{ID}/hidrive", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "ID"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid id in url: %d", chi.URLParam(r, "ID")), 400)
			return
		}

		r.ParseMultipartForm(10 << 20)

		modtime := r.Form.Get("modtime")
		name := r.Form.Get("name")

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, errors.Wrap(err, "error:").Error(), 500)
			return
		}
		defer file.Close()

		fmt.Printf("header: %s, name: %s, modtime: %v\n", header.Filename, name, modtime)
		meta, err := hdclient.UploadFile(fmt.Sprintf("bilder/%d", id), file, name, modtime)
		if err != nil {
			http.Error(w, errors.Wrap(err, "error:").Error(), 500)
			return
		}
		fmt.Printf("meta: %#v\n", meta)

		_, err = repo.InsertFoto(
			id, 0, meta.Name, int(meta.Size), meta.Name,
			"format", 1024, 768, time.Now(), "Caption", "kommentar",
		)
		if err != nil {
			http.Error(w, errors.Wrap(err, "error:").Error(), 500)
			return
		}
		// err = repo.SaveAusstellung(ausstellung)
		// if err != nil {
		// 	http.Error(w, fmt.Sprintf("Couldn‘t save ausstellung: %v", ausstellung), 500)
		// }
		// ausstellung, err = repo.LoadAusstellung(id)
		bytes, err := json.MarshalIndent(meta, "", "    ")
		if err != nil {
			w.Write([]byte(errors.Wrap(err, "error:").Error()))
			return
		}
		w.Write(bytes)
	})

	return r
}

func SerienList(db *Repo) ([]Serie, error) {

	serien, err := db.LoadSerien()
	if err != nil {
		return nil, err
	}

	return serien, nil
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

	err = GenerateThumbnail100(file, fotoID, config.Width, config.Height, media+"/thumbs/100/")
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

func GenerateThumbnail100(file io.ReadSeeker, id, width, height int, prefix string) error {

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
