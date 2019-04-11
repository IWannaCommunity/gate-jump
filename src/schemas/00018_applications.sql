CREATE TABLE `applications`(
    id INT(8) NOT NULL UNIQUE AUTO_INCREMENT,
    str_id VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    type VARCHAR(32) NOT NULL,
    secret VARCHAR(256) UNIQUE,
    redirect_uri VARCHAR(256),
    PRIMARY KEY (id)
)
