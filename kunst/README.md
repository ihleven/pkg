# pkg kunst


##  api

Handler eingeängt unter wolfgang-ihle.de/api
Interner Zugriff auf:
* repository
* hidrive drive

### /ausstellungen

* GET 
Liste der Ausstellungen

* POST /ausstellugnen
Neue Ausstellung anlegen

* GET /ausstellungen/{id}
Daten zu einer bestimmten Ausstellung (mit id)

* PUT /ausstellungen/{id}
Daten zu einer bestimmten Ausstellung ändern

* POST /ausstellungen/{id}/documents
Dateien zu einer bestimmten Ausstellung ins hidrive hochladen

* GET /ausstellungen/{id}/documents
-> in GET /ausstellungen/{id} integrieren

## Serien

* GET /serien
Liste aller Serien

* POST /serien
Neue Serie anlegen


* GET /serien/{ID}
Alle Daten zur Serie mit id ID bekommen

* POST /serien/{ID}
Daten zur Serie mit id ID ändern


## Bilder

* GET /bilder
Liste aller Bilder bekommen
 * Parameter: phase, sortby


* POST /bilder
Neue Serie anlegen


* GET /serien/{ID}
Alle Daten zur Serie mit id ID bekommen

* POST /serien/{ID}
Daten zur Serie mit id ID ändern





