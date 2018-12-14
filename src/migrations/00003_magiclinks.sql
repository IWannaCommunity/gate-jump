CREATE TABLE magic (
    id INT NOT NULL AUTO_INCREMENT,
    userid INT NOT NULL,
    magic TEXT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (userid) REFERENCES users(id)
)
