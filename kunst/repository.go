package kunst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/dbscan"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/ihleven/errors"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewRepo(database string) (*Repo, error) {

	ctx := context.Background()
	dbpool, err := pgxpool.Connect(ctx, fmt.Sprintf("postgresql://%s@localhost:5432/%s", database, database))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to database: %q", database)
	}

	repo := Repo{
		ctx:    ctx,
		dbpool: dbpool,
	}

	return &repo, nil
}

type Repo struct {
	ctx    context.Context
	dbpool *pgxpool.Pool
}

func (r *Repo) Select(dst interface{}, query string, args ...interface{}) error {

	err := pgxscan.Select(r.ctx, r.dbpool, dst, query, args...)
	if err != nil {
		// if errors.As(err, &pgx.ErrNoRows) {
		// 	return errors.NewWithCode(errors.NotFound, "Not found: %s", query)
		// }
		return errors.Wrap(err, "Error")
	}
	return nil
}

func (r *Repo) LoadAusstellungen() ([]Ausstellung, error) {

	stmt := "SELECT * FROM ausstellung ORDER BY id ASC"

	var ausstellungen []Ausstellung
	err := r.Select(&ausstellungen, stmt)
	if err != nil {
		return nil, err
	}

	return ausstellungen, nil
}

func (r *Repo) InsertAusstellung(a *Ausstellung) (int, error) {

	stmt := "INSERT INTO ausstellung (titel, untertitel, typ, jahr, von, bis, ort, venue, kommentar) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"

	row := r.dbpool.QueryRow(r.ctx, stmt, a.Titel, a.Untertitel, a.Typ, a.Jahr, a.Von, a.Bis, a.Ort, a.Venue, a.Kommentar)
	var returnid int
	err := row.Scan(&returnid)
	return returnid, errors.Wrap(err, "Could not insert ausstellung %v", a)
}

func (r *Repo) SaveAusstellung(a *Ausstellung) error {
	fmt.Println("ausstellung:", a)
	stmt := `
		UPDATE ausstellung SET titel=$1, untertitel=$2, typ=$3, jahr=$4, von=$5, bis=$6, ort=$7, venue=$8, kommentar=$9 WHERE id=$10
	`
	_, err := r.dbpool.Exec(r.ctx, stmt, a.Titel, a.Untertitel, a.Typ, a.Jahr, a.Von, a.Bis, a.Ort, a.Venue, a.Kommentar, a.ID)
	return errors.Wrap(err, "Could not save ausstellung %v", a)
}

// LoadAusstellung ...
func (r *Repo) LoadAusstellung(id int) (*Ausstellung, error) {

	stmt := "SELECT * FROM ausstellung WHERE id=$1"
	fmt.Println("LoadAusstellung")

	ausstellung := Ausstellung{Bilder: make([]Bild, 0), Fotos: make([]Foto, 0)}
	err := pgxscan.Get(r.ctx, r.dbpool, &ausstellung, stmt, id)

	if dbscan.NotFound(err) {
		fmt.Println(dbscan.NotFound(err))
		return nil, errors.NewWithCode(404, "not row found")
	} else if err != nil {
		return nil, errors.Wrap(err, "db error")
	}

	err = r.Select(&ausstellung.Bilder, "SELECT b.* FROM bild b, enthalten e WHERE b.id=e.bild_id AND ausstellung_id=$1", id)
	if err != nil {
		return nil, errors.Wrap(err, "db error bilder")
	}

	// err = r.Select(&ausstellung.Fotos, "SELECT * FROM foto WHERE aust_id=$1", id)
	// if err != nil {
	// 	return nil, err
	// }
	return &ausstellung, nil
}

func (r *Repo) LoadSerien() ([]Serie, error) {

	stmt := "SELECT * FROM serie ORDER BY id ASC"

	var serien []Serie
	err := r.Select(&serien, stmt)
	if err != nil {
		return nil, err
	}

	return serien, nil
}

func (r *Repo) InsertSerie(s *Serie) (int, error) {

	stmt := "INSERT INTO serie (slug, titel, untertitel, jahr, jahrbis, technik, traeger, hoehe, breite, tiefe, anmerkungen, kommentar, phase) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id"

	row := r.dbpool.QueryRow(r.ctx, stmt, s.Titel, s.Titel, s.Untertitel, s.Jahr, s.JahrBis, s.Technik, s.Träger, s.Höhe, s.Breite, s.Tiefe, s.Anmerkungen, s.Kommentar, s.Schaffensphase)
	var returnid int
	err := row.Scan(&returnid)
	return returnid, errors.Wrap(err, "Could not insert serie %v", s)
}

func (r *Repo) UpdateSerie(id int, fieldmap map[string]interface{}) (*Serie, error) {

	i := 2
	var fields []string
	values := []interface{}{id}
	for field, value := range fieldmap {
		if field == "id" {
			continue
		}
		values = append(values, fmt.Sprintf("%v", value))
		field := fmt.Sprintf("%s=$%d", field, i)
		fields = append(fields, field)
		i++
	}

	stmt := fmt.Sprintf("UPDATE serie SET %s WHERE id=$1", strings.Join(fields, ","))

	_, err := r.dbpool.Exec(r.ctx, stmt, values...)
	if err != nil {
		return nil, errors.Wrap(err, "Could not update serie %d => %v", id, fieldmap)
	}
	serie, loaderr := r.LoadSerie(id)
	if loaderr != nil {
		return serie, loaderr
	}
	return serie, nil
}

// LoadSerie ...
func (r *Repo) LoadSerie(id int) (*Serie, error) {

	stmt := "SELECT * FROM serie WHERE id=$1"

	serie := Serie{Bilder: make([]Bild, 0), Fotos: make([]Foto, 0)}
	err := pgxscan.Get(r.ctx, r.dbpool, &serie, stmt, id)
	if err != nil {
		return nil, err
	}

	err = r.Select(&serie.Bilder, "SELECT * FROM bild WHERE serie=$1", serie.Titel)
	if err != nil {
		return nil, err
	}

	// err = r.Select(&serie.Fotos, "SELECT * FROM foto WHERE aust_id=$1", id)
	// if err != nil {
	// 	return nil, err
	// }
	return &serie, nil
}

func (r *Repo) LoadBilder(phase, orderBy string) ([]Bild, error) {

	stmt := "SELECT * FROM bild"
	if phase != "" {
		stmt += " WHERE phase='" + phase + "'"
	}
	if orderBy != "" {
		stmt += " ORDER BY '" + orderBy + "' DESC"
	} else {
		stmt += " ORDER BY id DESC"
	}

	var bilder []Bild
	err := r.Select(&bilder, stmt)
	if err != nil {
		return nil, err
	}
	indexByID := make(map[int]int)
	for i, bild := range bilder {
		indexByID[bild.ID] = i
	}
	fmt.Println("bilder:", bilder)
	var fotos []Foto
	err = r.Select(&fotos, "SELECT * FROM foto WHERE id IN (SELECT foto_id FROM bild)")
	if err != nil {
		return nil, err
	}

	for i, foto := range fotos {
		if index, ok := indexByID[foto.BildID]; ok {
			bilder[index].IndexFoto = &fotos[i]
		}
	}
	// for i, _ := range bilder {
	// 	fmt.Println(i, bilder[i].IndexFoto)
	// }
	// fmt.Println(indexByID)
	return bilder, nil
}

func (r *Repo) LoadBild(id int) (*Bild, error) {

	stmt := "SELECT * FROM bild WHERE id=$1"

	bild := Bild{Fotos: make([]Foto, 0)}
	err := pgxscan.Get(r.ctx, r.dbpool, &bild, stmt, id)
	if err != nil {
		return nil, err
	}

	err = r.Select(&bild.Fotos, "SELECT * FROM foto WHERE bild_id=$1", id)
	if err != nil {
		return nil, err
	}
	return &bild, nil
}

func (r *Repo) InsertBild(bild *Bild) (int, error) {

	stmt := "INSERT INTO bild (titel, jahr, technik, traeger, hoehe, breite, tiefe, flaeche, foto_id, anmerkungen, kommentar, ordnung, phase, teile) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id"

	row := r.dbpool.QueryRow(r.ctx, stmt, bild.Titel, bild.Jahr, bild.Technik, bild.Bildträger, bild.Höhe, bild.Breite, bild.Tiefe, bild.Fläche, bild.IndexFotoID, bild.Anmerkungen, bild.Kommentar, bild.Überordnung, bild.Schaffensphase, bild.Teile)
	var returnid int
	err := row.Scan(&returnid)
	return returnid, errors.Wrap(err, "Could not insert bild %v", bild)
}

func (r *Repo) SaveBild(id int, bild *Bild) error {

	stmt := "UPDATE bild set  titel=$2, jahr=$3, technik=$4, traeger=$5, hoehe=$6, breite=$7, tiefe=$8, flaeche=$9, foto_id=$10, anmerkungen=$11, kommentar=$12, ordnung=$13, phase=$14, teile=$15 where id=$1"
	fmt.Println("saeBild:", bild.Schaffensphase)
	i, err := r.dbpool.Exec(r.ctx, stmt, id, bild.Titel, bild.Jahr, bild.Technik, bild.Bildträger, bild.Höhe, bild.Breite, bild.Tiefe, bild.Fläche, bild.IndexFotoID, bild.Anmerkungen, bild.Kommentar, bild.Überordnung, bild.Schaffensphase, bild.Teile)
	if err != nil {
		return err
	}
	fmt.Println(bild.Titel)

	fmt.Println("i:", i, err)
	return err
}

func (r *Repo) InsertFoto(id, index int, name string, size int, path, format string, width, height int, taken time.Time, caption, kommentar string) (int, error) {

	stmt := "INSERT INTO foto (bild_id,index,name,size,uploaded,path,format,width,height,taken,caption,kommentar) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id"

	row := r.dbpool.QueryRow(r.ctx, stmt, id, index, name, size, time.Now(), path, format, width, height, taken, caption, kommentar)
	var returnid int
	err := row.Scan(&returnid)
	bild, err := r.LoadBild(id)
	if err == nil && bild.IndexFotoID == 0 {
		bild.IndexFotoID = returnid
		r.SaveBild(id, bild)
	}

	return returnid, errors.Wrap(err, "Could not insert foto %v", id)

}

func (r *Repo) LoadFoto(id int) (*Foto, error) {

	stmt := "SELECT * FROM foto WHERE id=$1"

	var foto Foto
	err := pgxscan.Get(r.ctx, r.dbpool, &foto, stmt, id)
	if err != nil {
		return nil, err
	}

	return &foto, nil
}

func (r *Repo) LoadFotos() ([]Foto, error) {

	var fotos []Foto
	err := r.Select(&fotos, "SELECT * FROM foto")
	if err != nil {
		return nil, err
	}

	return fotos, nil
}

func (r *Repo) Update(table string, id int, fieldmap map[string]interface{}) error {

	i := 2
	var fields []string
	values := []interface{}{id}
	for field, value := range fieldmap {
		if field == "id" {
			continue
		}
		values = append(values, fmt.Sprintf("%v", value))
		field := fmt.Sprintf("%s=$%d", field, i)
		fields = append(fields, field)
		i++
	}

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE id=$1", table, strings.Join(fields, ","))

	_, err := r.dbpool.Exec(r.ctx, stmt, values...)
	return errors.Wrap(err, "Could not update %s %d => %v", table, id, fieldmap)
}

func (r *Repo) Delete(table string, key string, id int) error {
	if id == 0 {
		return errors.New("id 0 not supported")
	}

	stmt := fmt.Sprintf("DELETE FROM %s WHERE %s=$1", table, key)

	_, err := r.dbpool.Exec(r.ctx, stmt, id)
	return errors.Wrap(err, "Could not delete %s %d", table, id)
}
