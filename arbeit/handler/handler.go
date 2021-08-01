package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// type Marshaler interface {
// 	Render(bool) ([]byte, error)
// }

type params struct {
	year, month, week, day int
}

func parseURL(urlpath string) *params {

	var params params
	var regex = regexp.MustCompile(`\/(\d{4})\/?(kw|KW|Kw|w|W)?(\d{1,2})?\/?(\d{1,2})?`)
	matches := regex.FindStringSubmatch(urlpath)
	if matches != nil {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			params.year = year
		}
		if month, err := strconv.Atoi(matches[3]); err == nil {
			if matches[2] != "" {
				params.week = month
			} else {
				params.month = month
			}
		}
		if day, err := strconv.Atoi(matches[4]); err == nil {
			params.day = day
		}
	}
	fmt.Printf("%s => %+v\n", urlpath, params)
	return &params
}

// ArbeitHandler ...
func ArbeitHandler(uc *arbeit.Usecase) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		enableCors(&w)
		//sessionUser, _ := session.GetSessionUser(r, w)

		var response interface{}
		var err error

		params := parseURL(r.URL.Path)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")

		switch {
		case params.day != 0:
			// m, err = usecase.GetArbeitstag(params.year, params.month, params.day, 1)
			// m, err = ArbeitstagAction(w, r, params, usecase)

			switch r.Method {
			case http.MethodPut:

				var arbeitstag arbeit.Arbeitstag
				e := json.NewDecoder(r.Body).Decode(&arbeitstag)
				if e != nil {
					err = errors.NewWithCode(400, "Couldn‘t decode request: %v", e)
					break
				}

				response, err = uc.UpdateArbeitstag(params.year, params.month, params.day, 1, &arbeitstag)
				if err != nil {
					err = errors.Wrap(err, "Error updating arbeitstag %v", params)
					break
				}
				fallthrough

			case http.MethodGet:

				response, err = uc.GetArbeitstag(params.year, params.month, params.day, 1)
				if err != nil {
					err = errors.Wrap(err, "could not read arbeitstag %d/%d/%d %d", params.year, params.month, params.day, 1)
				}

			default:
				err = errors.NewWithCode(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
			}

		case params.week != 0:
			//err = params.HandleArbeitWoche(w, r, params)

		case params.month != 0:
			response, err = uc.GetArbeitsmonat(params.year, params.month, 1)

		case params.year != 0:
			response, err = uc.Arbeitsjahr(1, params.year)

		default:
			response, err = uc.ListArbeitsjahre(1)
		}

		fmt.Println("Get Arbeitstag", response, err)

		if err != nil {
			code := errors.Code(err)
			http.Error(w, err.Error(), int(code))
		} else {

			var bytes []byte
			err := encoder.Encode(response)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	}
}

// func ArbeitstagAction(w http.ResponseWriter, r *http.Request, p *params, usecase *arbeit.Usecase) (Marshaler, error) {
// 	switch r.Method {
// 	case http.MethodPut:

// 		var body arbeit.Arbeitstag
// 		err := json.NewDecoder(r.Body).Decode(&body)
// 		if err != nil {
// 			return nil, errors.NewWithCode(errors.BadRequest, "Could not decode request body")
// 		}
// 		arbeitstag, err := usecase.UpdateArbeitstag(p.year, p.month, p.day, 1, &body)
// 		fmt.Println("--- ArbeitstagAction", p, r.Method, arbeitstag, err)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "Could not update arbeitstag")
// 		}
// 		// return arbeitstag, nil
// 		fallthrough

// 	case http.MethodGet:

// 		arbeitstag, err := usecase.GetArbeitstag(p.year, p.month, p.day, 1)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "could not read arbeitstag %d/%d/%d %d", p.year, p.month, p.day, 1)
// 		}
// 		return arbeitstag, nil
// 		// js, err := json.MarshalIndent(arbeitstag, "", "\t")
// 		// if err != nil {
// 		// 	return nil, errors.Wrap(err, "could not marshal arbeitstag %d/%d/%d %d", p.year, p.month, p.day, 1)
// 		// }

// 		// w.Header().Set("Content-Type", "application/json")
// 		// w.Write(js)

// 	default:

// 		return nil, errors.NewWithCode(errors.BadRequest, "Only get or put")
// 		// enableCors(&w)
// 	}
// 	return nil, nil
// }

// func handleArbeitswoche(w http.ResponseWriter, r *http.Request, p *params) {
// 	arbeitswoche, err := arbeit.RetrieveArbeitsWoche(params.year, params.week, 1)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 	}
// 	json.NewEncoder(w).Encode(arbeitswoche)
// }
