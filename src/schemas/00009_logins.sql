CREATE TABLE logins (
    id INT NOT NULL AUTO_INCREMENT,
    userid INT NOT NULL,
    useruuid VARCHAR(255) NOT NULL,
    type VARCHAR(10) NOT NULL,
    token TEXT NOT NULL,
    expires SMALLINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (userid) REFERENCES users(id),
    FOREIGN KEY (useruuid) REFERENCES users(uuid)
)
