package kunst

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ihleven/errors"
	"github.com/jackc/pgx/v4/pgxpool"
)

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

func NewRepo() (*Repo, error) {

	ctx := context.Background()
	dbpool, err := pgxpool.Connect(ctx, "postgresql://mi@localhost:5432/mi")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to database: %q", "mi")
	}

	repo := Repo{
		ctx:    ctx,
		dbpool: dbpool,
	}

	return &repo, nil
}

func (r *Repo) LoadBilder() ([]Bild, error) {

	stmt := "SELECT id,jahr,name,technik,material,format,breite,hoehe,flaeche,foto_id FROM bilder ORDER BY id"

	var bilder []Bild
	err := r.Select(&bilder, stmt)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	fmt.Println("bilder:", bilder)

	return bilder, nil
}

func (r *Repo) LoadBild(id int) (*Bild, error) {

	stmt := "SELECT id,jahr,name,technik,material,format,breite,hoehe,flaeche,foto_id FROM bilder WHERE id=$1"

	var bild Bild
	err := pgxscan.Get(r.ctx, r.dbpool, &bild, stmt, id)
	if err != nil {
		return nil, err
	}

	err = r.Select(&bild.Fotos, "SELECT id, path, caption, width, height FROM fotos WHERE bild_id=$1", id)
	if err != nil {
		return nil, err
	}
	return &bild, nil
}

func (r *Repo) InsertFoto(id int, path, caption string, width, height int) (int, error) {

	stmt := "INSERT INTO fotos (bild_id,path,caption,width,height) VALUES ($1,$2,$3,$4,$5) RETURNING id"

	row := r.dbpool.QueryRow(r.ctx, stmt, id, path, caption, width, height)
	var returnid int
	err := row.Scan(&returnid)
	return returnid, errors.Wrap(err, "Could not insert bild %v", id)

}
