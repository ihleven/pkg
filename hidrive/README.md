# HiDrive

## media

Unter /media/ ist ein hidrive-Drive für den (lesenden) Zugriff auf den share/dir /wolfgang-ihle gemountet


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

### hd client

* 
* GetMeta
* GetDir
* 




### Client

client für die HiDriveApi. handhabt Kommunikation mit Api




Es gibt Funktionen für Dir, File, Meta, die Pfade nehmen und ReadCloaser zurückliefern


Handler kann dann entweder den Stream direkt zurückliefern oder nochmals parsen und 