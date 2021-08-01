package arbeit

import (
	"fmt"
	"strings"
	"time"

	"github.com/ihleven/pkg/kalender"
)

// Job beschreibt eine Entitaet, fuer die Arbeitszeit erfasst werden soll.
type Job struct {
	Code        string    `json:"code"`
	Account     int       `json:"account"`
	Nr          int       `json:"nr"`
	Arbeitgeber string    `json:"arbeitgeber"`
	Eintritt    time.Time `json:"eintritt"`
	Austritt    time.Time `json:"austritt"`
}

type Urlaub struct {
	Account int    `json:"account,omitempty"`
	Job     string `json:"job,omitempty"`
	Jahr    int    `json:"jahr"`

	Nr  int       `json:"nr"`
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`

	Urlaubstage       float64 `db:"num_urlaub" json:"urlaubstage"`
	Ausgleichstage    float64 `db:"num_ausgl" json:"ausgleichstage"`
	Sonderurlaubstage float64 `db:"num_sonder" json:"sonderurlaubstage"`

	Grund     string    `json:"grund"`
	Beantragt time.Time `json:"beantragt"`
	Genehmigt time.Time `json:"genehmigt"`
	Kommentar string    `json:"kommentar"`
}

type Arbeitsjahre []Arbeitsjahr

type Arbeitsjahr struct {
	Account int    `json:"account,omitempty"`
	Job     string `json:"job,omitempty"`
	Jahr    int    `json:"jahr"`

	Von *time.Time `db:"von" json:"von,omitempty"`
	Bis *time.Time `db:"bis" json:"bis,omitempty"`

	Arbeitstage    int `db:"ARBTG" json:"-"`
	Krankheitstage int `db:"K"     json:"-"`
	Bürotage       int `db:"B"     json:"-"`
	Homeoffice     int `db:"H"     json:"-"`
	Dienstreise    int `db:"D"     json:"-"`
	Zeitausgleich  int `db:"ZA"    json:"-"`

	Soll          float64 `json:"soll"`
	Ist           float64 `json:"ist"`
	Diff          float64 `json:"diff"`
	Saldo         float64 `json:"saldo"`
	Zeiterfassung float64 `json:"zeiterfassung"`

	Stat   ArbeitstageStat `json:"arbeitstage,omitempty"`
	Urlaub UrlaubStat      `json:"urlaub,omitempty"`

	Monate  []Arbeitsmonat `json:"monate,omitempty"`
	Urlaube []Urlaub       `json:"urlaube,omitempty"`
}

type UrlaubStat struct {
	Vorjahr      float64 `db:"uvorj" json:"vorjahr"`
	Anspruch     float64 `db:"uansp" json:"anspruch"`
	Tage         float64 `db:"utage" json:"tage"`
	Sonderurlaub float64 `db:"usond" json:"sonderurlaub"`
	Geplant      float64 `db:"ugepl" json:"geplant"`
	Rest         float64 `db:"urest" json:"rest"`
	Auszahlung   float64 `db:"uausz" json:"auszahlung"`
}

type ArbeitstageStat struct {
	Arbeitstage   int `db:"ARBTG" json:"arbeitstage"`
	Urlaub        int `db:"K"     json:"urlaubstage"`
	Krank         int `db:"K"     json:"krankheitstage"`
	Büro          int `db:"B"     json:"bürotage"`
	Homeoffice    int `db:"H"     json:"tageHomeoffice"`
	Dienstreise   int `db:"D"     json:"dienstreisetage"`
	Zeitausgleich int `db:"ZA"    json:"ausgleichstage"`
}

type Arbeitsmonat struct {
	// account integer NOT NULL,
	// job     varchar(8) NOT NULL,
	// jahr    integer NOT NULL,
	Monat int `json:"monat,omitempty"`

	Arbeitstage    int `db:"a" json:"arbeitstage"`
	Krankheitstage int `db:"k" json:"krankheitstage"`
	Urlaubstage    int `db:"u" json:"urlaubstage"`

	Soll          float64      `          json:"soll"`
	Ist           float64      `          json:"ist"`
	Diff          float64      `db:"diff" json:"diff"`
	Saldo         float64      `          json:"saldo"`
	Zeiterfassung float64      `          json:"zeiterfasssung"`
	Tage          []Arbeitstag `          json:"tage,omitempty"`
}

// type Arbeitswoche struct {
// 	Jahr int
// 	Nr   int
// 	//Job                   *Job
// 	Arbeitstage []Arbeitstag
// }

type Arbeitstag struct {
	Account int    `json:"account"` //domain.Account
	Datum   Date   `db:"datum" json:"datum"`
	Job     string `json:"job"`

	kalender.Tag `json:"tag"`
	// Jahr  int16 `json:"jahr,omitempty"`
	// Monat uint8 `json:"monat,omitempty"`

	// Typ: A-Arbeitstag, A2-Halber Arbeitstag, U-Urluab, F-Feiertag,
	Status ArbeitstagStatus `db:"status" json:"status,omitempty"`
	// B-Büro, D-Dienstreise, H-Homeoffice, K-Krankheit, U-Urlaub, Z-Zeitausgleich
	Kategorie ArbeitstagKategorie `db:"kategorie" json:"kategorie,omitempty"`
	//Krankheit string              `json:"krankmeldung,omitempty"`
	Urlaubstage float64 `json:"urlaubstage,omitempty"`

	Soll      float64    `json:"soll,omitempty"`
	Start     *time.Time `db:"start" json:"beginn"` //084300 => 8h43:00
	Ende      *time.Time `db:"ende" json:"ende"`    //173000 => 17h30:00
	Brutto    float64    `json:"brutto"`            //099700
	Pausen    float64    `json:"pausen"`
	Extra     float64    `json:"extra"`
	Netto     float64    `json:"netto"`          // Brutto + Extra - Pausen
	Differenz float64    `db:"diff" json:"diff"` // Netto - Soll

	// ergibt sich aus Saldo Vortag + Differenz
	Saldo     float64 `json:"saldo"`
	Kommentar string  `json:"kommentar"`
	// Saldo        *float64
	Zeitspannen []Zeitspanne ` json:"zeitspannen,omitempty"`
}

// type Datum struct {
// 	Datum    Date    `json:"date,omitempty"`
// 	Jahr     int16   `db:"jahr" json:"year,omitempty"`
// 	Monat    uint8   `db:"monat" json:"month,omitempty"`
// 	Tag      uint8   `db:"tag" json:"day,omitempty"`
// 	Jahrtag  uint16  `db:"jahrtag" json:"jahrtag,omitempty"`
// 	KwJahr   int16   `db:"kw_jahr" json:"kw_jahr,omitempty"`
// 	KwNr     uint8   `db:"kw" json:"kw_nr,omitempty"`
// 	KwTag    uint8   `db:"kw_tag" json:"kw_tag,omitempty"`
// 	Feiertag *string `json:"feiertag,omitempty"`

// 	// Ordinal int    `json:"ord,omitempty"`
// 	//monatsname string
// 	//tagesname  string
// }

// type Arbeitstag2 struct {
// 	Datum  Datum
// 	Typ    *string //
// 	Status *string //
// 	//Zeitspannen []Zeitspanne
// }

type Zeitspanne struct {
	// Job     string `json:"job"`
	Account int              `db:"account" json:"account"`
	Datum   Date             `db:"datum" json:"datum"`
	Nr      int              `db:"nr" json:"nr"`
	Status  ZeitspanneStatus `db:"status" json:"status"`
	Start   *time.Time       `db:"start" json:"start"`
	Ende    *time.Time       `db:"ende" json:"ende"`
	Dauer   float64          `db:"dauer" json:"dauer"`
	// Arbeitszeit float64          `json:"arbeitszeit"`
	// Netto       float64          `json:"netto"`
	//Kategorie    string              // AZ, Pause, Weg
	// Titel, Story, Beschreibung, Grund *string
}

type Date time.Time

const ctLayout = "2006-01-02"

// UnmarshalJSON Parses the json string in the custom format
func (ct *Date) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(ctLayout, s)
	*ct = Date(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct Date) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *Date) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(ctLayout))
}
