package repository

import (
	"fmt"
	"time"

	"github.com/ihleven/pkg/arbeit"
)

func (r *Repository) SetupJobs() error {

	var jobs = []arbeit.Job{
		arbeit.Job{"DBIS", 1, 1, "ALU", DateFromID(20010820), DateFromID(20071031)},
		arbeit.Job{"UKLFR", 1, 2, "Uniklinik Freiburg", DateFromID(20080401), DateFromID(20130131)},
		arbeit.Job{"AVE", 1, 3, "Averbis GmbH", DateFromID(20130201), DateFromID(20180731)},
		arbeit.Job{"IC", 1, 4, "Inter Chalet", DateFromID(20180801), DateFromID(20201231)},
	}
	for _, job := range jobs {
		err := r.SetupJob(job)
		eintrittsjahr := job.Eintritt.Year()
		austrittsjahr := job.Austritt.Year()
		for jahr := eintrittsjahr; jahr <= austrittsjahr; jahr++ {
			var von, bis *time.Time
			startmonat := 1
			endemonat := 12
			if jahr == eintrittsjahr {
				von = &job.Eintritt
				startmonat = int(job.Eintritt.Month())
			}
			if jahr == austrittsjahr {
				bis = &job.Austritt
				endemonat = int(job.Austritt.Month())
			}
			err = r.SetupArbeitsjahr(job.Account, job.Code, jahr, von, bis)
			if err != nil {
				fmt.Println(err)
			}

			for monat := startmonat; monat <= endemonat; monat++ {
				err = r.SetupArbeitsmonat(job.Account, job.Code, jahr, monat)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}
