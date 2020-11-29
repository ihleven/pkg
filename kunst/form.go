package kunst

import (
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/ihleven/errors"
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

func parseFormSubmitSerie(r *http.Request) (*Serie, error) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}
	var serie struct {
		Serie
		Von, Bis, Typ, Name string
		NumBilder           int
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	err = decoder.Decode(&serie.Serie, r.PostForm)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding form")
	}
	// if t, err := time.Parse("2006-01-02", ausstellung.Von); err == nil {
	// 	ausstellung.Ausstellung.Von = &t
	// }
	// if t, err := time.Parse("2006-01-02", ausstellung.Bis); err == nil {
	// 	ausstellung.Ausstellung.Bis = &t
	// }
	return &serie.Serie, nil
}
