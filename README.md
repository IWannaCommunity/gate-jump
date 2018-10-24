# gate-jump
Central Authentication Service for Delfruit, IWM, and other fangame community services!

## Environment Setup

* Install the MySQL server (v14.14 Distrib 5.7.24 for Linux x86_64): https://www.mysql.com/downloads/
* Install MySQL Driver for Go: `go get ggithub.com/go-sql-driver/mysql`
* Install Gorilla-Mux (golang http server framework) `go get github.com/gorilla/mux`
* Install jwt-go (jwt token encoder and decoder) `go get github.com/dgrijalva/jwt-go`
* Log into MySql by opening a terminal and typing `mysql -u root -p` and then entering the root password.
* Type `CREATE DATABASE gatejump` then `USE gatejump` and then past the following create table code below.

```sql
CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    password CHAR(60) BINARY NOT NULL,
    email VARCHAR(100),
    country CHAR(2),
    locale VARCHAR(20),
    date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    admin BOOL NOT NULL DEFAULT FALSE,
    verified BOOL NOT NULL DEFAULT FALSE,
    banned BOOL NOT NULL DEFAULT FALSE,
    last_token BLOB,
    last_login DATETIME,
    last_ip VARCHAR(50),
    PRIMARY KEY (id)
)
```

For more details on the `users` columns please look at `user.go` in the main directory of the project.

*Create a file and appropriately name the following information under `~/config/config.json`

```json
{
    "host":"0.0.0.0",
    "port":"80",
    "sslPort":"443",
    "database":{
        "username":"root",
        "password":"password",
        "dsn":"gatejump"
    },
    "https":{
        "certFile":"",
        "keyFile":""
    }
}
```




* Alternatively, import the user file from delfruit: ask Klazen for this! (Out of Date?)
```sql
select id as id
    , name as username
    , phash2 as password
    , email as email
    , locale as country
    , date_created as dateCreated
    , 0 as verified
    , banned as banned
    , is_admin as admin
    , '' as lastToken
    , date_last_login as lastLogin
    , last_ip as lastIP
from User where phash2 is not null and banned = 0;
```

## How to run

To build and run the project:

```shell
go build && gate-jump.exe
```

Navigate to `http://localhost:10420/`

If you see `{"alive": true}`, you're all set!

Try `http://localhost:10420/user/1` to see your user!


#Database Properties
