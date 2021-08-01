# cloud11-api

# Juni 2020

Es gibt eine Router fuer Drive und Arbeit, der in Main unter einer Basisroute (oder als Fallback) direkt beim zentralen Dispatcher eingehängt wird.
Also: unter /hidrive wird ein Treiber für HiDrive eingehängt
Dieser bekommt einene Pfad übergeben und liefert einen Serializer und einen Fehler zurück.
main.go

hidrive := hidrive.New(configuration)



# package httperror

erfuellt error interface und ermoeglicht in der applikation zeitnah fehelrcodes zu definieren
erkennt diverse standard fehlertypen
httperror.GetStatus(err) liefert http code und detailiert message entweder direkt aus einem httperror oder aus normaelem fehler (status 500)


The action takes HTTP requests (URLs and their methods) and uses that input to interact with the domain, after which it passes the domain's output to one and only one responder.
The domain can modify state, interacting with storage and/or manipulating data as needed. It contains the business logic.
The responder builds the entire HTTP response from the domain's output which is given to it by the action.


Interfaces:
 * Responder -> Respond(http.Responsewriter, http.Reuquest, interface{}) kann anhand des requests unterschiede machen
 * Renderer -> Render(http.Responsewriter, interface{}) eryeugt ausgabe auf w fuer interface{}

 * noch zu definierendes interface fuer domain logic

Actioneer ist struct, das http.HAndler erfuellt 

Drive ist interface, Hidrive und FS sind implemtierungen davon 


DriveHandler und DriveServeHAndler sind structs mit Drive und REsponder als komponenten
{
    drive Drive
    responder Responder
}



Der Webserver leitet mit dem ShiftPathDispatcher einen Request an den DriveHandler weiter.

## Handler 
DriveHandler erfüllt das HandlerInterface und bekommt von main einen DriveInteractor gewired.

Der Handler erüllt folgende Aufgabe:
* Alles was mit HTTP zu tun hat
* Authentifizierung
* Session
* CORS

Ruft Interactor mit Pfad, User auf und bekommt DriveObjekt und Error zurück
* DriveObjekt enhält alle Infos zum Pfad (Metadaten, Kinder, Breadcrumbs, ...)
* Error ist nil im Erfolgsfall und hat sonst einen Stacktrace, und detaillierte Fehlerbeschreibung. Außerdem kann ein HTTP-Statuscode abegeleitet werden.

Jetzt muss noch die REsponse erzeugt werden. Dazu erfüllt das DriveObjekt das Interface Responder, in dem mit der Methode
Respond() oder SerializeJSON je nach HTTP-Request die Rückgabe erzeugt wird, die anschließend über den Responswriter ausgegeben wird.
Responder Interface hat Methode, die MediaType übergeben bekommt und wenn JSON angefordert wird (Header, GET-Param), die Meta-Daten 

## Interactor


