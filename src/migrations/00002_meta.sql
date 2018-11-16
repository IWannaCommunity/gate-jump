CREATE TABLE meta (
    db_version SMALLINT UNSIGNED NOT NULL DEFAULT 0
);

INSERT INTO meta ( db_version ) VALUES ( 2 );
