package kalender

import (
	"fmt"
	"time"
)

func setupFeiertage(jahr int) map[int]string {

	feiertageMap := make(map[int]string)

	type feiertagGenerator func(int) Feiertag

	feiertage := []feiertagGenerator{Neujahr, Epiphanias, Valentinstag, Josefitag, Weiberfastnacht, Rosenmontag, Fastnacht, Aschermittwoch,
		Palmsonntag, Gründonnerstag, Karfreitag, Karsamstag, Ostern, Ostermontag, BeginnSommerzeit, Walpurgisnacht,
		TagDerArbeit, Staatsfeiertag, InternationalerTagDerPressefreiheit, Florianitag, TagDerBefreiung, Muttertag, ChristiHimmelfahrt,
		Pfingsten, Pfingstmontag, Dreifaltigkeitssonntag, Fronleichnam, InternationalerKindertag, TagDesMeeres, Weltflüchtlingstag,
		MariäHimmelfahrt, Rupertitag, TagDerDeutschenEinheit, TagDerVolksabstimmung, Erntedankfest, Nationalfeiertag,
		Reformationstag, Halloween, BeginnWinterzeit, Allerheiligen, Allerseelen, Martinstag, Karnevalsbeginn, Leopolditag, Weltkindertag,
		BußUndBettag, Thanksgiving, Blackfriday, Volkstrauertag, Nikolaus, MariäUnbefleckteEmpfängnis, MariäEmpfängnis,
		Totensonntag, ErsterAdvent, ZweiterAdvent, DritterAdvent, VierterAdvent, Heiligabend, Weihnachten, Christtag, ZweiterWeihnachtsfeiertag, Stefanitag, Silvester,
		InternationalerTagDesGedenkensAnDieOpferDesHolocaust, InternationalerFrauentag}
	for _, generator := range feiertage {
		feiertag := generator(jahr)
		feiertageMap[feiertag.YearDay()] = feiertag.Text
	}
	return feiertageMap
}

func ListKalendertage(year int) []Tag {

	feiertageMap := setupFeiertage(year)

	kalendertage := []Tag{}
	// feiertage := []feiertage.Feiertag{feiertage.Neujahr(year), feiertage.Valentinstag(year), feiertage.Epiphanias(year), feiertage.Weiberfastnacht(year)}
	// feiertageMap := make(map[int]string)
	// for _, feiertag := range feiertage {

	// 	feiertageMap[feiertag.Time.YearDay()] = feiertag.Text
	// }

	mez, _ := time.LoadLocation("Europe/Berlin")
	for d := time.Date(year, time.January, 1, 0, 0, 0, 0, mez); d.Year() == year; d = d.AddDate(0, 0, 1) {

		weekyear, week := d.ISOWeek()
		weekday := d.Weekday()
		if weekday == 0 {
			// Sonntag
			weekday = 7
		}
		kalendertag := Tag{
			ID:      id(d.Date()),
			Datum:   d,
			Jahr:    int16(d.Year()),
			Monat:   uint8(d.Month()),
			Tag:     uint8(d.Day()),
			Jahrtag: uint16(d.YearDay()),
			KwJahr:  int16(weekyear),
			KwNr:    uint8(week),
			KwTag:   uint8(weekday),
		}
		if feiertag, ok := feiertageMap[d.YearDay()]; ok {
			kalendertag.Feiertag = feiertag
		}
		fmt.Println(d, kalendertag)
		kalendertage = append(kalendertage, kalendertag)
	}
	return kalendertage
}

func id(year int, month time.Month, day int) int {
	return year*10000 + int(month)*100 + day
}
