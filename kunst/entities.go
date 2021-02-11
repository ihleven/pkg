// Package kunst is a virtual exhibition or collection of wolfgang ihle's work
package kunst

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ihleven/pkg/hidrive"
)

type Schaffensphase string

const (
	Früh                 Schaffensphase = "Frühwerk"
	Natur1                              = "Natur I Landschaft, Figur"
	Natur2                              = "Natur II Abstraktion"
	Entgegenständlichung                = "Entgegenständlichung"
	Monochrom                           = "Monochrome Malerei"
)

func (p Schaffensphase) IsValid() bool {
	switch p {
	case Früh, Natur1, Natur2, Entgegenständlichung, Monochrom:
		return true
	}
	return false
}

type Ausstellung struct { // Exhibition

	ID         int        `db:"id"           json:"id"`
	Code       string     `db:"code"           json:"code"`
	Ort        string     `db:"ort"          json:"ort"`
	Jahr       int        `db:"jahr"         json:"jahr"`
	Venue      string     `db:"venue"        json:"venue"`
	Titel      string     `db:"titel"        json:"titel"`
	Untertitel string     `db:"untertitel"   json:"untertitel"`
	Typ        string     `db:"typ"          json:"typ"` // einzel, sammel, dauerleihgabe, ....
	Von        *time.Time `db:"von"          json:"von"`
	Bis        *time.Time `db:"bis"          json:"bis"`
	Kommentar  string     `db:"kommentar"    json:"kommentar"`
	// NumBilder int        `db:"num_bilder"          json:"num_bilder"  schema:"num_bilder"`
	Bilder    []Bild      `db:"-"    json:"bilder"`
	Fotos     []Foto      `db:"-"    json:"fotos"`
	Dokumente interface{} `db:"-"    json:"dokumente"`
}

type Katalog struct {
	ID         int `db:"id"           json:"id"`
	Code       string
	Titel      string
	Untertitel string
	Jahr       int
	Datum      *time.Time
	Kommentar  string `db:"kommentar"    json:"kommentar"`

	Modified *time.Time `db:"modified" json:"modified"`

	IndexFoto *hidrive.Meta  `db:"-"        json:"foto"`
	Fotos     []hidrive.Meta `db:"-"        json:"fotos"`
	Dokumente interface{}    `db:"-"        json:"dokumente"`
}

type Serie struct {
	ID             int    `db:"id"         json:"id"`
	Slug           string `db:"slug"       json:"slug,omitempty"  `
	Jahr           int    `db:"jahr"       json:"jahr,omitempty"  `
	Titel          string `db:"titel"      json:"titel"           `
	Untertitel     string `db:"untertitel" json:"untertitel"      `
	Anzahl         int    `db:"anzahl"    json:"anzahl,omitempty"`
	JahrBis        int    `db:"jahrbis"    json:"jahrbis,omitempty"`
	Technik        string `db:"technik"    json:"technik"         `                 // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Träger         string `db:"traeger" schema:"traeger"   json:"traeger"         ` // Leinwand, Papier, Holz,
	Höhe           int    `db:"hoehe"      schema:"hoehe" json:"hoehe"           `  //
	Breite         int    `db:"breite"     json:"breite"`                           //
	Tiefe          int    `db:"tiefe"      json:"tiefe"`                            //
	Anmerkungen    string `db:"anmerkungen" json:"anmerkungen"`                     // Anmerkungen des Künstlers
	Kommentar      string `db:"kommentar"  json:"kommentar"       `
	Schaffensphase string `db:"phase"       json:"phase" schema:"phase"` // Natur
	Bilder         []Bild `db:"-"          json:"bilder"           schema:"-"`
	Fotos          []Foto `db:"-"          json:"fotos"            schema:"-"`
}

type Bild struct {
	ID          int     `db:"id"          json:"id"` //
	Directory   string  `db:"dir"        json:"dir"`
	Jahr        int     `db:"jahr"        json:"jahr"`  //
	Titel       string  `db:"titel"       json:"titel"` //
	SerieID     *int    `db:"serie_id"       json:"serie_id"`
	Serie       *Serie  `db:"-"       json:"serie"`
	SerieNr     int     `db:"serie_nr"    json:"serie_nr"`
	Technik     string  `db:"technik"     json:"technik"`                    // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Bildträger  string  `db:"traeger"     json:"traeger" schema:"traeger"`   // Leinwand, Papier, Holz,
	Höhe        int     `db:"hoehe"       json:"hoehe"       schema:"hoehe"` //
	Breite      int     `db:"breite"      json:"breite"`                     //
	Tiefe       int     `db:"tiefe"       json:"tiefe"`                      //
	Fläche      float64 `db:"flaeche"     json:"flaeche"`                    // Bildfläche in qm
	Anmerkungen string  `db:"anmerkungen" json:"anmerkungen"`                // Anmerkungen des Künstlers
	Kommentar   string  `db:"kommentar"   json:"kommentar"`                  // Kommentare zum Bild, nicht für die Öffentlichkeit gedacht
	// Überordnung    string  `db:"ordnung"     json:"ueberordnung"`               //
	Schaffensphase string `db:"phase"       json:"phase" schema:"phase"` // Natur
	Fotos          []Foto `db:"-"           json:"fotos" `
	IndexFotoID    int    `db:"foto_id"     json:"foto_id" schema:"foto_id"` //
	IndexFoto      *Foto  `db:"foto"     json:"foto" schema:"-"`             //
	// Systematik     string `db:"-"    json:"sytematik"`    //
	// Ordnung        string `db:"-"    json:"ordnung"`      //
	// Hauptfoto   *Foto
	// Diptychon / Triptychon
	// AusstellungID int
	// KatalogID     int
	Teile    int        `db:"teile"       json:"teile"`
	Modified *time.Time `db:"modified"    json:"modified"`
	// SeriePtr *Serie     `db:"-"       json:"Serie"`
}

type Foto struct {
	ID        int       `db:"id"        json:"id"`
	BildID    int       `db:"bild_id"   json:"-"`
	Bild      *Bild     `db:"-"         json:"-"`
	Index     int       `db:"index"     json:"nr"`
	Name      string    `db:"name"      json:"name"`
	Size      int       `db:"size"      json:"size"`
	Uploaded  time.Time `db:"uploaded"  json:"uploaded"`
	Path      string    `db:"path"      json:"path"`
	Format    string    `db:"format"    json:"format"`
	Width     int       `db:"width"     json:"width"`
	Height    int       `db:"height"    json:"height"`
	Taken     time.Time `db:"taken"     json:"taken"`
	Caption   string    `db:"caption"   json:"caption"`
	Kommentar string    `db:"kommentar" json:"kommentar"`
	Labels    []string  `db:"-"         json:"labels"`

	// AusstellungID *int `db:"aust_id"         json:"aust_id"`
	// KatalogID     int

	Serie *string
}

// Natur I Landschaft, Figur
// Natur II Abstraktion
// Entgegenständlichung
// Monochrome Malerei

type Dokument struct {
	ID   int
	Pfad string // austellungen/9/RedeHelmut.pdf

}

func render(data interface{}, w http.ResponseWriter) {
	bytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(bytes)
}
