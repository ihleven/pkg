# HiDrive

## Drive-Handler

Handler, der Methoden und Pfade auf die HiDrive-Api abbildet. 
Über ein Cookie wird dem HAndler eine Authentifizierung zugeordnet: 
* username: Zuordnung to Struct{hidrive-Token, home-Verzeichnis, eventuell auch ein eListe mit Permissions}
* token: Zuordnung zu einem Hidrive-token
* {
    hitoken: "...",
    home: "...",
    permissions: ["/bilder", "/Halde/Super8"]
}

Einem Drive ist ein Pfad-Prefix zugeordent, der den verfügbaren Filebaum im Hidrive einschränkt:
* ein konstanter Pfad zu einem Verzeichnis (e.g. /user/matt.ihle/wolfgang-ihle), im Defaultfall ist das / 
* ein Slug /home wird auf das Home-Verzeichnis des angemeldeten Benuzters gemappt (e.g. /home/bilder -> /users/matt.ihle/bilder für angemeldeten User matt)

* dir     => GetDir auf prefix + restpfad
* meta    => GetMeta auf prefix + restpfad
* files   => GetFile auf pid
* file    => GetFile auf prefix + restpfad
* thumbs  => GetThumbnail auf prefix + restpfad
* default: wie files


## Drive

Hat Methoden, um Directories auszulesen, Dateien zu streamen oder hochzuladen:
```
type Drive interface{
    GetDir(path, username string) (*Dir, error)
    GetMeta(path, username string) (*Dir, error)
    MkDir
    UploadFile
}
```

### hidrive.Client

Zugriff auf die hidrive-Endpunkte
* Parameter spiegeln Endpunkt-Parameter 1:1
* bekommt Zugiffstokens übergeben
* gibt die hidrive-Endpunkt-Rückgabe 1:1 weiter
* GetDir / PostDir / DeleteDir
* GetFile




### Client

client für die HiDriveApi. handhabt Kommunikation mit Api




Es gibt Funktionen für Dir, File, Meta, die Pfade nehmen und ReadCloser zurückliefern


Handler kann dann entweder den Stream direkt zurückliefern oder nochmals parsen und 

### FS

Implementiert io.FS interface

```
type FS interface{
    Open(name string) (fs.File, error)
    ReadDir(name string) ([]fs.DirEntry, error)
    ReadFile(name string) ([]byte, error)
}
```

Eigenschaften: Pfad-Prefix und Token.
Damit werden Open, Stat, ReadDir, etc. mittels des enthalteten hidrive.Client implementiert.
Eine FS ist nicht flexibel einsetzbar, sondern wird in  jedem Request aus Prefix und Token erzeugt.
Prefix kann statisch sein, oder das home des angemeldeten Users sein. Token über angemeldeten User oder entpunkt-statisch. 

```
type FS struct {
	client *HiDriveClient
	prefix string
	token  string // *Token2
}
```



### Drive 

Wrapper um Client, der Benutzerauthentifizierung und Pfad-Berechnungen und Zugriffsrechte beisteuert.

drive.Mkdir(name, authkey)
drive.Rmdir(name, authkey)
drive.Rm(name, authkey)
drive.Lsdir(name, authkey)
drive.GetMeta(name, authkey)
drive....

### Drive-Handler

Ist ein Weblaufwerk, das spezifische Pfade in Client-Calls übersetzt:

* /dir/{path} => hidrive.GetDir
* /meta/{path} => hidrive.GetMeta
* /thumbs/{id} => hidrive.GetThumb
* /files/{id} => hidrive.GetFile
* /{path} => hidrive.GetFile ?!?

Rückgabe ist JSON bzw. Rohdaten

Der angemeldete User wird aus dem Request geholt. 
Es kann dabei für User eine Konfiguration hinterlegt sein.
Pfadmapping:
 * home wird durch users/<username> ersetzt
 * /prefix wird vor den Pfad gehängt
Tokenzuordnung:
 * dem authuser wird ein hidrive.Token zugeordnet, beispielsweise eines für lediglich öffenltihchen Zugriff
ACL:
 * Es kann eine Glob-Liste mit erlaubten / verboteten Pfade enthalten sein.

### Fileserver

/hiserve/{path}

Falls {path} auf eine Datei zeigt, wird diese binär ausgeliefert, falls es sich um ein Verzeichnis handelt, wird 
der Inhalt von index.html im Verzeichnis ausgeliefert, bzw. 404 

### Template-Handler


Dieser Handler liefert für alle Pfade in einem Template gerenderte Metadaten aus.
Die Pfade haben keinen filespezifischen Prefix, sondern es wird aus dem Pfad ermittlelt ob es sich um eine Datei oder um ein Verzeichnis handelt.
Templates können dabei für Filetypen registiert werden. Dir und File sind Pflicht, für spezielle Filetypen können aber eigene Templates hinterlegt werden (Image, Markdown)





