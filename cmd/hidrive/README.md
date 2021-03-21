# hidrive cmd

## Pfade

Absolute Pfade werden 1:1 ins hidrive übernommen: /users/matt.ihle/akutell oder /wolfgang-ihle/bilder/43

Relative Pfade werden immer auf ein Eltern-Verzeichnis im HiDrive bezogen. Dieses kann über das Flag --parent mit einem absoluten Pfad explizit festgelegt werden.
Ohne --parent-Flag wird das Home-Verzeichnis des angemeldeten Benutzers als Default angenommen. d.h. obiger absoute Pfad ist gleichwertig mit aktuell (wenn matt.ihle angemeldet ist).

~-Pfade, also beispielsweise ~aktuell oder ~/aktuell werden auf das Home-Verzeichnis des angemeldeten Benutzers erweitert.

## Commands

### login

Anmeldung eines Benutzers mit seinem HiDrive-Account

### logout

Abmelden des angemeldeten Benutzers

### list 

### info

