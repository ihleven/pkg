package repository

import (
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"
	pq "github.com/lib/pq"
)

func (r *Repository) SetupJob(job arbeit.Job) error {

	stmt := `
		INSERT INTO c11_job (code, account, nr, arbeitgeber, eintritt, austritt) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.DB.Exec(stmt, job.Code, job.Account, job.Nr, job.Arbeitgeber, job.Eintritt, job.Austritt)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { //"23505": "unique_violation"
		err = r.SaveJob(job)
	}
	if err != nil {
		return errors.Wrap(err, "Could not setup job %v", job)
	}
	return nil
}

func (r *Repository) SaveJob(job arbeit.Job) error {

	stmt := `
		UPDATE c11_job
		   SET account=$2, nr=$3, arbeitgeber=$4, eintritt=$5, austritt=$6
		 WHERE code=$1 
	`
	_, err := r.DB.Exec(stmt, job.Code, job.Account, job.Nr, job.Arbeitgeber, job.Eintritt, job.Austritt)
	if err != nil {
		return errors.Wrap(err, "Could not update c11_job %v", job)
	}
	return nil
}
