
DROP TABLE foto;
DROP TABLE bild;
DROP TABLE serie;

CREATE TABLE serie (
    id          integer PRIMARY KEY,
    jahr        integer NOT NULL DEFAULT 0,
    titel       text    NOT NULL DEFAULT '',
    num_bilder integer NOT NULL DEFAULT 0
);


CREATE TABLE bild (
    id          serial  PRIMARY KEY,
    jahr        integer NOT NULL DEFAULT 0,
    titel       text    NOT NULL DEFAULT '',
    serie       text    NOT NULL DEFAULT '',
    serie_nr    integer NOT NULL DEFAULT 0,
    technik     text    NOT NULL DEFAULT '',
    traeger     text    NOT NULL DEFAULT '',
    hoehe       integer NOT NULL DEFAULT 0,
    breite      integer NOT NULL DEFAULT 0,
    tiefe       integer NOT NULL DEFAULT 0,
    flaeche     double precision NOT NULL DEFAULT 0.0,
    foto_id     integer DEFAULT 0,
    -- hauptfoto   integer DEFAULT 0,
   
    anmerkungen text NOT NULL DEFAULT '',
    kommentar   text NOT NULL DEFAULT '',
    ordnung     text NOT NULL DEFAULT '',
    phase       text NOT NULL DEFAULT ''
);




CREATE TABLE foto (
    id        serial     PRIMARY KEY,
    bild_id   integer     REFERENCES bild(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    index     integer     NOT NULL DEFAULT 0,

    name      text        NOT NULL DEFAULT '',
    size      integer     NOT NULL DEFAULT 0,
    -- type     text NOT NULL DEFAULT '',

    uploaded  timestamptz NOT NULL DEFAULT Now(),
    path      text        NOT NULL DEFAULT '',
    format    text        NOT NULL DEFAULT '',
    width     integer     NOT NULL DEFAULT 0,
    height    integer     NOT NULL DEFAULT 0,
    taken     timestamptz NOT NULL,
    caption   text        NOT NULL DEFAULT '',
    kommentar text        NOT NULL DEFAULT ''
);



