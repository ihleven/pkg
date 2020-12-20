package kunst

// import (
// 	"fmt"
// 	"image"
// 	"io"
// 	"io/ioutil"
// 	"mime"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/disintegration/imageorient"
// 	"github.com/disintegration/imaging"
// 	"github.com/ihleven/errors"
// )

// r.Post("/bilder/{ID}/fotos", func(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "ID")
// 	foto, err := UploadFoto(repo, id, r)
// 	fmt.Println("foto", foto, err)
// 	bytes, err := json.MarshalIndent(foto, "", "    ")
// 	if err != nil {
// 		w.Write([]byte(errors.Wrap(err, "error:").Error()))
// 	}
// 	w.Write(bytes)
// })

// func UploadFoto(db *Repo, idstr string, r *http.Request) (*Foto, error) {

// 	// w.Header().Set("Access-Control-Allow-Origin", "*")
// 	// w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,PATCH,OPTIONS")
// 	// w.Header().Set("Access-Control-Allow-Headers", "*")

// 	r.ParseMultipartForm(10 << 20)

// 	bildID, err := strconv.Atoi(r.Form.Get("id"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Printf("File upload for bild %v --- %s\n", bildID, idstr)

// 	file, header, err := r.FormFile("photo")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	name, err := save(file, header, bildID, media)
// 	if err != nil {
// 		return nil, err
// 	}

// 	config, format, err := getConfig(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println("config:", config, format, err)

// 	// TRANSAKTION: cancel wenn bild nicht gespeichert werden kann
// 	fotoID, err := db.InsertFoto(
// 		bildID, 0, header.Filename, 0, strings.TrimPrefix(name, media+"/"),
// 		format, config.Width, config.Height, time.Now(), "Caption", "kommentar",
// 	)
// 	if err != nil {
// 		fmt.Println("insert error:", err)

// 		return nil, err
// 	}
// 	fmt.Println("fotoID:", fotoID)

// 	err = generateThumbnail(file, fotoID, media+"/thumbs/")
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = GenerateThumbnail100(file, fotoID, config.Width, config.Height, media+"/thumbs/100/")
// 	if err != nil {
// 		return nil, err
// 	}

// 	foto, err := db.LoadFoto(fotoID)
// 	return foto, err
// }

// func save(file io.ReadSeeker, header *multipart.FileHeader, id int, path string) (string, error) {
// 	var f *os.File
// 	var err error
// 	nconflict := 0
// 	ts := time.Now().Format("060102-150405")
// 	for i := 0; i < 10000; i++ {
// 		name := fmt.Sprintf("%s/%d-%s-%s", path, id, ts, header.Filename)
// 		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
// 		if os.IsExist(err) {
// 			if nconflict++; nconflict > 10 {
// 				fmt.Println("nconflict:", nconflict)
// 			}
// 			continue
// 		}
// 		defer f.Close()
// 		break
// 	}

// 	bytes, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		return "", err
// 	}

// 	n, err := f.Write(bytes)
// 	if err != nil {
// 		return "", err
// 	}
// 	if n != int(header.Size) {
// 		return "", errors.New("written bytes and upload file size differ: %v != %v", n, header.Size)
// 	}

// 	return f.Name(), nil
// }

// func saveTemp(file io.ReadSeeker, header *multipart.FileHeader, id int, path string) (string, error) {
// 	suffixes, err := mime.ExtensionsByType(header.Header["Content-Type"][0])
// 	suffix := suffixes[0]

// 	tempfile, err := ioutil.TempFile(path, fmt.Sprintf("%d-*%s", id, suffix))
// 	if err != nil {
// 		return "", err
// 	}
// 	defer tempfile.Close()

// 	bytes, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		return "", err
// 	}

// 	n, err := tempfile.Write(bytes)
// 	if err != nil {
// 		return "", err
// 	}
// 	if n != int(header.Size) {
// 		return "", errors.New("written bytes and upload file size differ: %v != %v", n, header.Size)
// 	}

// 	return tempfile.Name(), nil
// }

// func getConfig(file io.ReadSeeker) (*image.Config, string, error) {
// 	if _, err := file.Seek(0, 0); err != nil {
// 		return nil, "", err
// 	}
// 	config, format, err := imageorient.DecodeConfig(file)
// 	if err != nil {
// 		errors.Wrap(err, "error in DecodeConfig")
// 	}
// 	fmt.Printf("format: %v\n", format)
// 	return &config, format, nil
// }

// func generateThumbnail(file io.ReadSeeker, id int, prefix string) error {

// 	if _, err := file.Seek(0, 0); err != nil {
// 		return err
// 	}
// 	img, _, err := image.Decode(file)
// 	if err != nil {
// 		return err
// 	}
// 	var dst *image.NRGBA

// 	dst = imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
// 	path := fmt.Sprintf("%s%d.png", prefix, id)
// 	err = imaging.Save(dst, path)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func GenerateThumbnail100(file io.ReadSeeker, id, width, height int, prefix string) error {

// 	if _, err := file.Seek(0, 0); err != nil {
// 		return err
// 	}
// 	// img, _, err := image.Decode(file)
// 	img, _, err := imageorient.Decode(file)
// 	if err != nil {
// 		return err
// 	}

// 	b := img.Bounds()
// 	width = b.Max.X
// 	height = b.Max.Y

// 	// fmt.Println("width = ", width, height)

// 	if width < height {
// 		height = 100 * height / width
// 		width = 100
// 	} else {
// 		width = 100 * width / height
// 		height = 100
// 	}
// 	// dst = imaging.Thumbnail(img, width, height, imaging.CatmullRom)
// 	dst := imaging.Resize(img, width, height, imaging.Lanczos)
// 	path := fmt.Sprintf("%s%d.png", prefix, id)
// 	fmt.Println(width, height, path)
// 	err = imaging.Save(dst, path)
// 	if err != nil {
// 		fmt.Println("saving error:", dst, err)

// 		return err
// 	}
// 	return nil
// }
