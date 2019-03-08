CREATE TABLE memberships (
    userid INT NOT NULL,
    groupid INT NOT NULL,
    FOREIGN KEY (userid) REFERENCES users(id),
    FOREIGN KEY (groupid) REFERENCES groups(id)
)
