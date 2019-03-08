CREATE TABLE permissions (
    groupid INT NOT NULL AUTO_INCREMENT,
    scopeid INT NOT NULL AUTO_INCREMENT,
    FOREIGN KEY (groupid) REFERENCES groups(id),
    FOREIGN KEY (scopeid) REFERENCES scopes(id),
)
