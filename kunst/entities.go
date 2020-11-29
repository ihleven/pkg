// Package kunst is a virtual exhibition or collection of wolfgang ihle's work
package kunst

import (
	"time"
)

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
	Bilder    []Bild     `db:"-"    json:"bilder"`
	Fotos     []Foto     `db:"-"    json:"fotos"`
	Dokumente []Dokument `db:"-"    json:"dokumente"`
}

type Katalog struct {
	ID    int
	Code  string
	Name  string
	Jahr  int
	Datum time.Time
}

type Serie struct {
	ID        int    `db:"id"        json:"id"`
	Jahr      int    `db:"jahr"      json:"jahr,omitempty"    schema:"jahr"`
	Titel     string `db:"titel"     json:"titel"             schema:"name"`
	Anzahl    int    `db:"anzahl"    json:"anzahl,omitempty"  schema:"num_bilder"`
	Kommentar string `db:"kommentar" json:"kommentar"         schema:"kommentar"`
	Bilder    []Bild `db:"-"         json:"bilder"             schema:"-"`
	Fotos     []Foto `db:"-"         json:"fotos"            schema:"-"`
}

type Bild struct {
	ID             int     `db:"id"          json:"id"`    //
	Jahr           int     `db:"jahr"        json:"jahr"`  //
	Titel          string  `db:"titel"       json:"titel"` //
	Serie          string  `db:"serie"       json:"serie"`
	SerieNr        int     `db:"serie_nr"    json:"serie_nr"`
	Technik        string  `db:"technik"     json:"technik"`                    // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Bildträger     string  `db:"traeger"     json:"traeger" schema:"traeger"`   // Leinwand, Papier, Holz,
	Höhe           int     `db:"hoehe"       json:"hoehe"       schema:"hoehe"` //
	Breite         int     `db:"breite"      json:"breite"`                     //
	Tiefe          int     `db:"tiefe"       json:"tiefe"`                      //
	Fläche         float64 `db:"flaeche"     json:"flaeche"`                    // Bildfläche in qm
	Anmerkungen    string  `db:"anmerkungen" json:"anmerkungen"`                // Anmerkungen des Künstlers
	Kommentar      string  `db:"kommentar"   json:"kommentar"`                  // Kommentare zum Bild, nicht für die Öffentlichkeit gedacht
	Überordnung    string  `db:"ordnung"     json:"ueberordnung"`               //
	Schaffensphase string  `db:"phase"       json:"phase"`                      // Natur
	Fotos          []Foto  `db:"-"           json:"fotos" `
	IndexFotoID    int     `db:"foto_id"     json:"foto_id" schema:"foto_id"` //
	IndexFoto      *Foto   `db:"foto"     json:"foto" schema:"-"`             //
	// Systematik     string `db:"-"    json:"sytematik"`    //
	// Ordnung        string `db:"-"    json:"ordnung"`      //
	// Hauptfoto   *Foto
	// Ditychon / Triptychon
	// AusstellungID int
	// KatalogID     int
	Teile    int        `db:"teile"       json:"teile"`
	Modified *time.Time `db:"modified"    json:"modified"`
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
}

// Natur I Landschaft, Figur
// Natur II Abstraktion
// Entgegenständlichung
// Monochrome Malerei

type Dokument struct {
	ID   int
	Pfad string // austellungen/9/RedeHelmut.pdf

}
