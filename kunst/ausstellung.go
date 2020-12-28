package kunst

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/schema"
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"
)

func parseFormSubmitAusstellung(r *http.Request, id int) (*Ausstellung, error) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}

	var ausstellung struct {
		Ausstellung
		Von, Bis string
	}
	decoder := schema.NewDecoder()
	// decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&ausstellung, r.PostForm)
	fmt.Println(" +++++++++++++ parseFormSubmitAusstellung", err)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}
	if t, err := time.Parse("2006-01-02", ausstellung.Von); err == nil {
		ausstellung.Ausstellung.Von = &t
	}
	if t, err := time.Parse("2006-01-02", ausstellung.Bis); err == nil {
		ausstellung.Ausstellung.Bis = &t
	}
	if id != 0 && ausstellung.ID != 0 && id != ausstellung.ID {
		return nil, errors.NewWithCode(400, "id in url (%s) and ausstellung (%d) differ", id, ausstellung.ID)
	}

	return &ausstellung.Ausstellung, nil
}

type AusstellungHandler struct {
	repo  *Repo
	drive *hidrive.Drive
}

func (a *AusstellungHandler) ListCreateAusstellungen(w http.ResponseWriter, r *http.Request, username string) error {

	if r.Method == "POST" {
		ausstellung, err := parseFormSubmitAusstellung(r, 0)

		id, err := a.repo.InsertAusstellung(ausstellung)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t insert ausstellung: %v", ausstellung)
		}

		_, err = a.drive.Mkdir("ausstellungen", strconv.Itoa(id), username)
		if err != nil && errors.Code(err) != 409 { // 409 wenn dir ex.
			return errors.Wrap(err, "Couldn‘t mkdir ausstellung/%d", id)
		}

		ausstellung, err = a.repo.LoadAusstellung(id)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t load data for created Ausstellung %d", id)
		}
		render(ausstellung, w)
		return nil
	}

	ausstellungen, err := a.repo.LoadAusstellungen()
	if err != nil {
		return errors.Wrap(err, "Couldn‘t load Ausstellungen")
	}
	if ausstellungen == nil {
		ausstellungen = []Ausstellung{}
	}
	render(ausstellungen, w)
	return nil
}

func (a *AusstellungHandler) GetUpdateDeleteAusstellung(r *http.Request, id int, username string) (*Ausstellung, error) {

	switch r.Method {
	case "GET":
		ausstellung, err := a.repo.LoadAusstellung(id)
		dir, err := a.drive.GetDir(fmt.Sprintf("/ausstellungen/%d", id), username)
		if err != nil {
			return nil, errors.Wrap(err, "error =>")
		}
		ausstellung.Dokumente = dir.Members

		return ausstellung, errors.Wrap(err, "Error loading ausstellung %v", id)

	case "PUT":
		ausstellung, err := parseFormSubmitAusstellung(r, id)
		err = a.repo.SaveAusstellung(id, ausstellung)
		fmt.Println("PUT ausstellung:", ausstellung, err)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t save ausstellung: %v", id)
		}
		return a.repo.LoadAusstellung(id)

	case "DELETE":
		err := a.repo.Delete("ausstellung", "id", id)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t delete ausstellung: %v", id)
		}

		err = a.drive.Rmdir("ausstellungen/"+strconv.Itoa(id), username)
		if errors.Code(err) == 404 {
			fmt.Printf("no folder for ausstellung %d found\n", id)
			return nil, nil
		}
		fmt.Println("error:", errors.Code(err), err)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t delete hidrive ausstellung folder %d", id)
		}
	}
	return nil, nil
}

func (a *AusstellungHandler) AusstellungDocuments(r *http.Request, id int, username string) (interface{}, error) {

	// Dateien zu einer bestimmten Ausstellung ins hidrive hochladen
	if r.Method == "POST" {

		r.ParseMultipartForm(10 << 20)

		modtime := r.Form.Get("modtime")
		name := r.Form.Get("name")

		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return a.drive.CreateFile(fmt.Sprintf("ausstellungen/%d", id), file, name, modtime, username)
	}

	dir, err := a.drive.GetDir(fmt.Sprintf("/ausstellungen/%d", id), username)
	if err != nil {
		if errors.Code(err) == 404 {
			return []string{}, nil
		}
		return nil, errors.Wrap(err, "error =>")
	}
	return dir.Members, nil
}
