DROP TABLE foto;
DROP TABLE enthalten;
DROP TABLE bild;
DROP TABLE serie;
DROP TABLE ausstellung;
DROP TABLE katalog;


CREATE TABLE katalog (
    id         serial  PRIMARY KEY,
    -- code       text    NOT NULL DEFAULT '', -- UNIQUE,      -- secondary key
    jahr       integer NOT NULL DEFAULT 0,
    titel      text    NOT NULL DEFAULT '',
    untertitel text    NOT NULL DEFAULT '',
    kommentar  text    NOT NULL DEFAULT ''
);


CREATE TABLE ausstellung (
    id         serial  PRIMARY KEY,
    -- code       text    NOT NULL DEFAULT '', -- UNIQUE,      -- secondary key
    titel      text    NOT NULL DEFAULT '',
    untertitel text    NOT NULL DEFAULT '',
    typ        text    NOT NULL DEFAULT '',
    jahr       integer NOT NULL DEFAULT 0,
    von        date    NULL,
    bis        date    NULL,
    ort        text    NOT NULL DEFAULT '',
    venue      text    NOT NULL DEFAULT 0,
    kommentar  text    NOT NULL DEFAULT ''
);


CREATE TABLE serie (
    id          serial  PRIMARY KEY,
    slug        text    NOT NULL UNIQUE,      -- secondary key
    jahr        integer NOT NULL DEFAULT 0,
    jahrbis     integer NOT NULL DEFAULT 0,
    titel       text    NOT NULL UNIQUE DEFAULT '',
    untertitel  text    NOT NULL UNIQUE DEFAULT '',
    anzahl      integer NOT NULL DEFAULT 0,
    technik     text    NOT NULL DEFAULT '',
    traeger     text    NOT NULL DEFAULT '',
    hoehe       integer NOT NULL DEFAULT 0,
    breite      integer NOT NULL DEFAULT 0,
    tiefe       integer NOT NULL DEFAULT 0,
    phase       text    NOT NULL DEFAULT ''
    anmerkungen text    NOT NULL DEFAULT ''
    kommentar   text    NOT NULL DEFAULT ''
);

CREATE TABLE bild (
    id          serial  PRIMARY KEY,
    jahr        integer NOT NULL DEFAULT 0,
    titel       text    NOT NULL DEFAULT '',
    serie       text    REFERENCES serie(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    serie_nr    integer NOT NULL DEFAULT 0,
    technik     text    NOT NULL DEFAULT '',
    traeger     text    NOT NULL DEFAULT '',
    hoehe       integer NOT NULL DEFAULT 0,
    breite      integer NOT NULL DEFAULT 0,
    tiefe       integer NOT NULL DEFAULT 0,
    flaeche     double precision NOT NULL DEFAULT 0.0,
    teile       integer NOT NULL DEFAULT 0,
    foto_id     integer NOT NULL DEFAULT 0,   
    anmerkungen text    NOT NULL DEFAULT '',
    kommentar   text    NOT NULL DEFAULT '',
    ordnung     text    NOT NULL DEFAULT '',
    phase       text    NOT NULL DEFAULT '',
    modified    timestamp with time zone not null default now()

);

CREATE TABLE enthalten (
    id             serial  PRIMARY KEY,
    bild_id        integer REFERENCES bild(id)        ON UPDATE CASCADE ON DELETE CASCADE,
    katalog_id     integer REFERENCES katalog(id)     ON UPDATE CASCADE ON DELETE CASCADE,
    ausstellung_id integer REFERENCES ausstellung(id) ON UPDATE CASCADE ON DELETE CASCADE
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
    -- kat_id    integer     REFERENCES katalog(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    -- aust_id   integer     REFERENCES ausstellung(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    serie_id  integer     REFERENCES serie(id) ON UPDATE CASCADE ON DELETE RESTRICT,
);

CREATE TABLE dokument (
    id        serial     PRIMARY KEY,
    format
    
);



