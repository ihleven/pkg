package kunst

import (
	"net/http"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"
)

func KatalogHandler(repo *Repo, drive *hidrive.Drive) func(http.ResponseWriter, *http.Request, int, string) error {

	list := func() ([]Katalog, error) {
		kataloge, err := repo.LoadKataloge()
		if err != nil {
			return nil, errors.Wrap(err, "Couldn‘t load Ausstellungen")
		}
		if kataloge == nil {
			kataloge = []Katalog{}
		}

		return kataloge, nil
	}

	retrieve := func(id int) (*Katalog, error) {
		katalog, err := repo.LoadKatalog(id)
		if err != nil {
			return nil, errors.Wrap(err, "error:")
		}
		return katalog, nil
	}

	return func(w http.ResponseWriter, r *http.Request, id int, authuser string) error {

		var err error
		var response interface{}

		switch id {
		case 0:
			switch r.Method {
			case "POST":
				// response, err = KatalogHandler.ListCreateAusstellung(r, username)

			case "GET":
				response, err = list()
			}
		default:
			switch r.Method {
			case "PUT":
			case "GET":
				response, err = retrieve(id)
			}

		}

		if err == nil {
			render(response, w)
		}
		return errors.Wrap(err, "")
	}
}
