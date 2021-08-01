package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"
	"github.com/ihleven/pkg/kalender"
)

func (r Repository) Zeitspannen() error {

	type Zeitspanne struct {
		ID           int
		Nr           int
		Typ          sql.NullString
		Von          *time.Time
		Bis          *time.Time
		Dauer        sql.NullFloat64
		Beschreibung sql.NullString
		ArbeitstagID int `db:"arbeitstag_id"`
		// Titel, Story, Grund *string
		// Arbeitszeit float64          `json:"arbeitszeit"`
	}

	// titel, story, grund immer leer, arbeitszeit immer false
	var query = "SELECT id, nr, typ, von, bis, dauer, beschreibung, arbeitstag_id FROM arbeits_zeitspanne"

	zeitspannen := []arbeit.Zeitspanne{}
	err := r.DB.Select(&zeitspannen, query)
	return err
}

func NullFloat(value sql.NullFloat64) float64 {
	if value.Valid {
		return value.Float64
	}
	return 0
}

func NullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}
func JobID(account, day int) string {
	if account == 1 {
		switch {
		case day >= 20010820 && day <= 20071031:
			return "DBIS"
		case day >= 20080401 && day <= 20130131:
			return "UKLFR"
		case day >= 20130201 && day <= 20180731:
			return "AVE"
		case day >= 20180801:
			return "IC"
		default:
			return ""
		}
	}
	return ""
}
func Jobs(account, jahr int) []string {
	if account == 1 {
		switch jahr {
		case 2013:
			return []string{"AVE"}
		case 2014:
			return []string{"AVE"}
		case 2015:
			return []string{"AVE"}
		case 2016:
			return []string{"AVE"}
		case 2017:
			return []string{"AVE"}
		case 2018:
			return []string{"AVE", "IC"}
		case 2019:
			return []string{"IC"}
		}
	}
	return []string{}
}
func DateFromID(id int) time.Time {

	return time.Date(id/10000, time.Month(id%10000/100), id%100, 0, 0, 0, 0, time.FixedZone("", 0))
}

func (r *Repository) LegacyArbeitsjahre() error {

	type LegacyArbeitsjahr struct {
		ID                    int
		Vorjahr               sql.NullFloat64 `db:"urlaub_vorjahr"`
		Anspruch              sql.NullFloat64 `db:"urlaub_anspruch"`
		Tage                  sql.NullFloat64 `db:"urlaub_tage"`
		Geplant               sql.NullFloat64 `db:"urlaub_geplant"`
		Rest                  sql.NullFloat64 `db:"urlaub_rest"`
		Soll                  sql.NullFloat64 `db:"soll"`
		Ist                   sql.NullFloat64 `db:"ist"`
		Diff                  sql.NullFloat64 `db:"differenz"`
		TageArbeit            sql.NullFloat64 `db:"tage_arbeit"`
		TageKrank             sql.NullFloat64 `db:"tage_krank"`
		TageFreizeitausgleich sql.NullFloat64 `db:"tage_freizeitausgleich"`
		TageBuero             sql.NullFloat64 `db:"tage_buero"`
		TageDienstreise       sql.NullFloat64 `db:"tage_dienstreise"`
		TageHomeoffice        sql.NullFloat64 `db:"tage_homeoffice"`
		TageFrei              sql.NullFloat64 `db:"tage_frei"`
		JahrID                int             `db:"jahr_id"`
		UserID                int             `db:"user_id"`
	}

	var query = `
		SELECT id, urlaub_vorjahr, urlaub_anspruch, urlaub_tage, urlaub_geplant, urlaub_rest, soll, ist, differenz, tage_freizeitausgleich, tage_krank, tage_arbeit, tage_buero, tage_dienstreise, tage_homeoffice, tage_frei, jahr_id, user_id
      	  FROM arbeitsjahr
	`

	arbeitsjahre := []LegacyArbeitsjahr{}
	err := r.DB.Select(&arbeitsjahre, query)

	for _, a := range arbeitsjahre {
		for _, job := range Jobs(a.UserID, a.JahrID) {
			arbeitsjahr := arbeit.Arbeitsjahr{
				Account: a.UserID,
				Job:     job,
				Jahr:    a.JahrID,
			}
			if a.Soll.Valid {
				arbeitsjahr.Soll = a.Soll.Float64
			}
			if a.Ist.Valid {
				arbeitsjahr.Ist = a.Ist.Float64
			}
			if a.Diff.Valid {
				arbeitsjahr.Diff = a.Diff.Float64
			}
			err = r.SaveArbeitsjahr(arbeitsjahr)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

	}
	return nil
}

func (r Repository) LegacyArbeitstage() error {

	type LegacyArbeitstag struct {
		ID int
		// Jahr  int16 `json:"jahr,omitempty"`
		// Monat uint8 `json:"monat,omitempty"`
		Status string // ArbeitstagStatus `db:"status" json:"status,omitempty"`
		Typ    sql.NullString
		// Kategorie         string
		Krank             bool
		Krankheit         sql.NullString  `json:"krankmeldung,omitempty"`
		Urlaubstage       sql.NullFloat64 `json:"urlaubstage,omitempty"`
		Freizeitausgleich sql.NullFloat64

		Soll      sql.NullFloat64 `json:"soll,omitempty"`
		Start     *time.Time      `db:"start" json:"beginn"` //084300 => 8h43:00
		Ende      *time.Time      `db:"ende" json:"ende"`    //173000 => 17h30:00
		Brutto    sql.NullFloat64 `json:"brutto"`            //099700
		Pausen    sql.NullFloat64 `json:"pausen"`
		Netto     sql.NullFloat64 `json:"netto"`               // Brutto + Extra - Pausen
		Differenz sql.NullFloat64 `db:"differenz" json:"diff"` // Netto - Soll

		Kommentar      sql.NullString
		Tagebuch       sql.NullString
		ArbeitsjahrID  int           `db:"arbeitsjahr_id"`
		ArbeitsmonatID int           `db:"arbeitsmonat_id"`
		ArbeitswocheID int           `db:"arbeitswoche_id"`
		DienstreiseID  sql.NullInt64 `db:"dienstreise_id"`
		TagID          int           `db:"tag_id"`
		UrlaubID       sql.NullInt64 `db:"urlaub_id"`
		UserID         int           `db:"user_id"`
	}

	var query = `
		SELECT id, status, typ, urlaubstage, freizeitausgleich, krank, krankheit, soll, start, ende, brutto, pausen, netto, differenz, kommentar, tagebuch, arbeitsjahr_id, arbeitsmonat_id, arbeitswoche_id, dienstreise_id, tag_id, urlaub_id, user_id
      	  FROM arbeitstag
	`

	arbeitstage := []LegacyArbeitstag{}
	err := r.DB.Select(&arbeitstage, query)

	for _, a := range arbeitstage {

		arbeitstag := arbeit.Arbeitstag{
			Account: a.UserID,
			Job:     JobID(a.UserID, a.ID/1000),
			// Jahr:        int16(a.ArbeitsjahrID / 1000),
			// Monat:       uint8(a.ArbeitsmonatID / 1000 % 100),
			Status:      arbeit.ArbeitstagStatus(a.Status),
			Kategorie:   arbeit.ArbeitstagKategorie(NullString(a.Typ)),
			Urlaubstage: NullFloat(a.Urlaubstage),
			Soll:        NullFloat(a.Soll),
			Start:       a.Start,
			Ende:        a.Ende,
			Brutto:      NullFloat(a.Brutto),
			Pausen:      NullFloat(a.Pausen),
			Extra:       0,
			Netto:       NullFloat(a.Netto),
			Differenz:   NullFloat(a.Differenz),
			Saldo:       0,
			Kommentar:   NullString(a.Kommentar),
		}
		err = r.SaveArbeitstag(a.UserID, DateFromID(a.TagID), JobID(a.UserID, a.TagID), arbeitstag)
		if err != nil {
			fmt.Println(err)
		}
	}
	return err
}

func (r Repository) MigrateKalendertage() ([]kalender.Tag, error) {

	type LegacyDatum struct {
		kalender.Tag
		Feiertag sql.NullString `db:"feiertag"`
	}

	var query = `
		SELECT id, jahr_id, monat, tag, datum, feiertag, kw_jahr, kw_nr, kw_tag, jahrtag, ordinal
      	  FROM kalendertag
	`
	// 		monatsname, tagesname, kalendermonat_id, kalenderwoche_id

	kalendertageLegacy := []LegacyDatum{}
	err := r.DB.Select(&kalendertageLegacy, query)
	if err != nil {
		return nil, err
	}

	kalendertageMigrated := []kalender.Tag{}
	for index, kalendertagLegacy := range kalendertageLegacy {
		id := kalendertagLegacy.ID
		jahr := id / 10000
		monat := id % 10000 / 100
		tag := id % 100
		date := time.Date(jahr, time.Month(monat), tag, 0, 0, 0, 0, time.FixedZone("", 0))
		kwyear, kw := date.ISOWeek()

		if !kalendertagLegacy.Datum.Equal(date) {
			return nil, errors.New("date diff: %v vs %v", kalendertagLegacy.Datum, date)
		}
		kalendertag := kalender.Tag{
			ID:       kalendertagLegacy.ID,
			Datum:    date,
			Jahr:     int16(jahr),
			Monat:    uint8(monat),
			Tag:      uint8(tag),
			Jahrtag:  uint16(date.YearDay()),
			KwJahr:   int16(kwyear),
			KwNr:     uint8(kw),
			KwTag:    uint8(date.Weekday()),
			Feiertag: "",
		}
		if kalendertagLegacy.Feiertag.Valid {
			kalendertag.Feiertag = kalendertagLegacy.Feiertag.String
		}
		err := r.UpsertKalendertag(kalendertag)
		if err != nil {
			return kalendertageMigrated, err
		}
		kalendertageMigrated = append(kalendertageMigrated, kalendertag)
		fmt.Println(index, kalendertag, err)
	}

	return kalendertageMigrated, nil
}
