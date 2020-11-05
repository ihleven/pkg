package kunst

import "time"

type Serie struct {
	Code string
	Name string
}

type Bild struct {
	ID           int       `db:"id"        json:"id"`                       //
	Name         string    `db:"name"      json:"name"`                     //
	Jahr         int       `db:"jahr"      json:"jahr"`                     //
	Datum        time.Time `db:"datum"     json:"datum"`                    //
	Technik      string    `db:"technik"   json:"technik"`                  //
	Material     string    `db:"material"  json:"material"`                 //
	Format       string    `db:"format"    json:"format"`                   //
	Breite       int       `db:"breite"    json:"b"`                        //
	Höhe         int       `db:"hoehe"     json:"h"`                        //
	Fläche       float64   `db:"flaeche"   json:"f"`                        // Bildfläche in qm
	IndexFotoID  int       `db:"foto_id"   json:"foto_id" schema:"foto_id"` //
	Fotos        []Foto    `db:"-"   json:"fotos" `
	Beschreibung string    `db:"beschreibung"   json:"beschreibung"`
	Kommentar    string    `db:"kommentar"   json:"kommentar"`
}

type Foto struct {
	ID   int   `db:"id"   json:"id"`
	Bild *Bild `db:"-"   json:"-"`
	// Index   bool   `db:"index"   json:"index"`
	Pfad    string `db:"path"   json:"path"`
	Caption string `db:"caption"   json:"caption"`
	Width   int    `db:"width"   json:"width"`
	Height  int    `db:"height"   json:"height"`
}
