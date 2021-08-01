package kalender

import (
	"time"
)

//     datum    |          feiertag          | kw_jahr | kw_nr | kw_tag | jahrtag | ordinal | monatsname | tagesname   | kalendermonat_id | kalenderwoche_id

type Tag struct {
	ID       int       `db:"id"       json:"id,omitempty"`
	Datum    time.Time `db:"tagdatum"    json:"date,omitempty"`
	Jahr     int16     `db:"jahr"  json:"year,omitempty"`
	Monat    uint8     `db:"monat"    json:"month,omitempty"`
	Tag      uint8     `db:"tag"      json:"day,omitempty"`
	Jahrtag  uint16    `db:"jahrtag"  json:"jahrtag,omitempty"`
	KwJahr   int16     `db:"kw_jahr"  json:"kw_jahr,omitempty"`
	KwNr     uint8     `db:"kw"    json:"kw_nr,omitempty"`
	KwTag    uint8     `db:"kw_tag"   json:"kw_tag,omitempty"`
	Ordinal  int       `db:"ordinal"  json:"ord,omitempty"`
	Feiertag string    `db:"feiertag"        json:"feiertag,omitempty"`
}
