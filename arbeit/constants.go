package arbeit

// ArbeitstagStatus ...
type ArbeitstagStatus string

// ArbeitstagStatus ...
const (
	StatusFrei       ArbeitstagStatus = "-"
	StatusArbeitstag ArbeitstagStatus = "A"
	StatusFeiertag   ArbeitstagStatus = "F"
	//	StatusBetriebsfrei     ArbeitstagStatus = "F"
)

// ArbeitstagKategorie ...
type ArbeitstagKategorie string

// Ohne, Krank
const (
	Ohne         ArbeitstagKategorie = "-"
	Krank        ArbeitstagKategorie = "K"
	Urlaubstag   ArbeitstagKategorie = "U"
	Sonderurlaub ArbeitstagKategorie = "S"

	Buero             ArbeitstagKategorie = "B"
	Homeoffice        ArbeitstagKategorie = "H"
	Dienstreise       ArbeitstagKategorie = "D"
	Freizeitausgleich ArbeitstagKategorie = "F"
)

// ZeitspanneKategorie ...
type ZeitspanneStatus string

// ZeitspanneKategorie ...
const (
	StatusArbeitszeit     ZeitspanneStatus = "A"
	StatusPause           ZeitspanneStatus = "P"
	StatusExtra           ZeitspanneStatus = "E"
	StatusWeg             ZeitspanneStatus = "W"
	StatusRestpausenabzug ZeitspanneStatus = "R"
)
