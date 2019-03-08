CREATE TABLE permissions (
    groupid INT NOT NULL,
    scopeid INT NOT NULL,
    FOREIGN KEY (groupid) REFERENCES groups(id),
    FOREIGN KEY (scopeid) REFERENCES scopes(id)
)
