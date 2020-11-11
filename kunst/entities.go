package kunst

import "time"

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
	Schaffensphase string  `db:"phase"       json:"phase"`                      //
	Fotos          []Foto  `db:"-"           json:"fotos" `
	IndexFotoID    int     `db:"foto_id"     json:"foto_id" schema:"foto_id"` //
	// Systematik     string `db:"-"    json:"sytematik"`    //
	// Ordnung        string `db:"-"    json:"ordnung"`      //
	// Hauptfoto   *Foto
	// Ditychon / Triptychon
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
}

type Katalog struct {
	ID    int
	Code  string
	Name  string
	Jahr  int
	Datum time.Time
}

type Serie struct {
	Code string
	Name string
}
