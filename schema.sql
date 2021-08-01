-- CREATE TABLE go_kalendertag  (
CREATE TABLE c11_datum (
    id       integer PRIMARY KEY,
    datum    date NOT NULL UNIQUE,
    jahr     integer NOT NULL,
    monat    integer NOT NULL,
    tag      integer NOT NULL,
    jahrtag  integer NOT NULL,
    kw_jahr  integer NOT NULL,
    kw       integer NOT NULL,
    kw_tag   integer NOT NULL,
    feiertag text NOT NULL DEFAULT '',
    UNIQUE (jahr, monat, tag),
    UNIQUE (kw_jahr, kw, kw_tag),
    UNIQUE (jahr, jahrtag)
);


CREATE TABLE c11_reise (
    -- id        | integer                |           | not null | 
    code      varchar(32) NOT NULL UNIQUE, 
    jahr      integer NOT NULL,
    monat     integer NOT NULL DEFAULT 0, 
    -- typ
    ziel      varchar(100) NOT NULL DEFAULT '', 
    von       date NOT NULL DEFAULT '1970-01-01', 
    bis       date NOT NULL DEFAULT '1970-01-01', 
    -- name      character varying(255)
    kommentar text NOT NULL DEFAULT '', 
    PRIMARY KEY (jahr,ziel)
);

CREATE TABLE account  (
    id integer PRIMARY KEY,
    -- uuid UUID UNIQUE DEFAULT gen_random_uuid(),
    -- role role NOT NULL DEFAULT 'guest',
    create_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);



DROP TABLE c11_zeitspanne;
DROP TABLE c11_arbeitstag;
DROP TABLE c11_arbeitsmonat;
DROP TABLE c11_dienstreise;
DROP TABLE c11_urlaub;
DROP TABLE c11_arbeitsjahr;
DROP TABLE c11_job;

CREATE TABLE c11_job  (
    code        varchar(8) PRIMARY KEY,
    account     integer NOT NULL,
    nr          integer NOT NULL,
    arbeitgeber varchar(32) NOT NULL DEFAULT '',
    eintritt    date NULL,
    austritt    date NULL,
    UNIQUE (code, account),
    UNIQUE (account, nr)
);




CREATE TABLE c11_arbeitsjahr  (
    account integer NOT NULL,
    job    varchar(8) NOT NULL, -- REFERENCES go_job()
    jahr  integer NOT NULL,

    von   date NULL,
    bis   date NULL,

    ARBTG int DEFAULT 0,  -- Arbeitstage
    K     int DEFAULT 0,  -- Krankheitstage
	B     int DEFAULT 0,  -- Bürotage
	H     int DEFAULT 0,  -- Homeoffice
    D     int DEFAULT 0,  -- Dienstreise
    ZA    int DEFAULT 0,  -- Zeitausgleich

    soll  float NOT NULL DEFAULT 0,
    ist   float NOT NULL DEFAULT 0,
    diff  float NOT NULL DEFAULT 0,
    saldo float NOT NULL DEFAULT 0,
    zeiterfassung float NOT NULL DEFAULT 0,

    uvorj float NOT NULL DEFAULT 0,
    uansp float NOT NULL DEFAULT 0,
    usond float NOT NULL DEFAULT 0,
    ugepl float NOT NULL DEFAULT 0,
    urest float NOT NULL DEFAULT 0,
    uausz float NOT NULL DEFAULT 0,
    
    PRIMARY KEY (account,job,jahr)
);

CREATE TABLE c11_urlaub  (
    account integer NOT NULL,
    job     varchar(8) NOT NULL, -- REFERENCES go_job()
    jahr    integer NOT NULL,

    nr   integer NOT NULL,
    von  date    NOT NULL,
    bis  date    NOT NULL,

    num_urlaub float NOT NULL,
    num_ausgl  float NOT NULL,
    num_sonder float NOT NULL,

    grund      text NOT NULL,
    beantragt  date NOT NULL,
    genehmigt  date NOT NULL,
    kommentar  text NOT NULL,

    PRIMARY KEY (account, job, jahr, nr),
    -- FOREIGN KEY (account) REFERENCES account (account),
    FOREIGN KEY (account,job) REFERENCES c11_job (account, code),
    FOREIGN KEY (account,job,jahr) REFERENCES c11_arbeitsjahr (account,job,jahr)
);


CREATE TABLE c11_dienstreise  (
    account integer NOT NULL,
    job     varchar(8) NOT NULL, -- REFERENCES go_job()
    jahr    integer NOT NULL,

    nr   integer NOT NULL,
    von  timestamp with time zone    NOT NULL DEFAULT '1970-01-01', 
    bis  timestamp with time zone    NOT NULL DEFAULT '1970-01-01', 
    -- id             | integer                  |           | not null | 
    ziel           text NOT NULL  DEFAULT '',
    grund          text NOT NULL  DEFAULT '',
    auslagen       float NOT NULL DEFAULT 0,
    erstattung     date NOT NULL DEFAULT '1970-01-01', 
    kommentar      text NOT NULL,
    -- arbeitsjahr_id | integer                  |           | not null | 
    -- reise_id       | integer                  |           |          | 
    -- user_id        | integer                  |           | not null | 
    PRIMARY KEY (account, job, jahr, nr),
    -- FOREIGN KEY (account) REFERENCES account (account),
    FOREIGN KEY (account,job) REFERENCES c11_job (account, code),
    FOREIGN KEY (account,job,jahr) REFERENCES c11_arbeitsjahr (account,job,jahr)
);
 
CREATE TABLE c11_arbeitsmonat  (
    account integer NOT NULL,
    job     varchar(8) NOT NULL,
    jahr    integer NOT NULL,
    monat   integer NOT NULL,

    A int DEFAULT 0,  -- Arbeitstage
    K int DEFAULT 0,  -- Krankheitstage
    U int DEFAULT 0,  -- Krankheitstage

    soll  float NOT NULL DEFAULT 0,
    ist   float NOT NULL DEFAULT 0,
    diff  float NOT NULL DEFAULT 0,
    saldo float NOT NULL DEFAULT 0,
    zeiterfassung float NOT NULL DEFAULT 0,

    PRIMARY KEY (account,job,jahr,monat),
    FOREIGN KEY (account,job,jahr) REFERENCES c11_arbeitsjahr (account,job,jahr)
);

CREATE TABLE c11_arbeitswoche  (
    account integer NOT NULL,
    job    varchar(8) NOT NULL,
    jahr  integer NOT NULL,
    kw   integer NOT NULL,

    A  int DEFAULT 0,  -- Arbeitstage
    K  int DEFAULT 0,  -- Krankheitstage
    U  int DEFAULT 0,  -- Krankheitstage

    soll  float NOT NULL DEFAULT 0,
    ist   float NOT NULL DEFAULT 0,
    diff  float NOT NULL DEFAULT 0,
    saldo float NOT NULL DEFAULT 0,
    zeiterfassung float NOT NULL DEFAULT 0,

    PRIMARY KEY (account, job, jahr, kw),
    FOREIGN KEY (account, job, jahr) REFERENCES c11_arbeitsjahr (account, job, jahr)
);

CREATE TABLE c11_arbeitstag  (
    account integer,
    datum   date NOT NULL,
    job     varchar(8) NOT NULL,
    
    jahr    integer NOT NULL,
    monat   integer NOT NULL,
    --woche integer NOT NULL,

    status       char(1) NOT NULL,
    kategorie    char(1) NOT NULL,
    -- krankmeldung | boolean                  |           | not null | false
    -- urlaubstage       NUMERIC(5, 2) NOT NULL,
    -- freizeitausgleich boolean,  -- | double precision
    -- krank             | boolean                  |           | not null | 
    -- krankheit         | text  
    soll         float NOT NULL,
    start        timestamp with time zone NULL,
    ende         timestamp with time zone NULL,
    brutto       float NOT NULL,
    pausen       float NOT NULL,
    extra        float NOT NULL,
    netto        float NOT NULL,
    diff         float NOT NULL,
    
    -- saldo        float NOT NULL,

    kommentar    text  NOT NULL,

    PRIMARY KEY (account, datum, job),
    FOREIGN KEY (account,job) REFERENCES c11_job (account,code),
    FOREIGN KEY (account,job,jahr) REFERENCES c11_arbeitsjahr (account,job,jahr)
);


CREATE TABLE c11_zeitspanne (
    account integer,
    datum   date                     NOT NULL,
    job     varchar(8) NOT NULL,
    
    nr      integer                  NOT NULL,

    status  varchar(1)               NOT NULL,
    start   timestamp with time zone,
    ende    timestamp with time zone,
    dauer   float                    NOT NULL DEFAULT 0,
    azfktor float NOT NULL DEFAULT 0,
    netto   float                    NOT NULL DEFAULT 0,

 -- titel         character varying(100) NOT NULL,
 -- story         character varying(100) NOT NULL,
    kommentar  text  NOT NULL DEFAULT '',
 -- grund         character varying(255)  NOT NULL,
 
    PRIMARY KEY (account, datum, nr),
    FOREIGN KEY (account, datum, job) REFERENCES c11_arbeitstag (account, datum, job)
);
              