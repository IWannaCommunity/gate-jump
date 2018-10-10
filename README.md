# gate-jump
Central Authentication Service for Delfruit, IWM, and other fangame community services!

## Environment Setup

* Install the MongoDB server: https://www.mongodb.com/download-center?jmp=nav#community
* Install MongoDB Driver for Go: `go get github.com/mongodb/mongo-go-driver/mongo`
* Install Gorilla-Mux (golang http server framework) `go get github.com/gorilla/mux`
* Get MongoDB Compass Community Edition (GUI for MongoDB): https://www.mongodb.com/download-center#compass
* Open Compass and create a database called `gatejump` and a collection called `users`, and add a user with userid 1:

```
userid:1 Int64
username:"Klazen108" String
password:<a hashed bcrypt password> String
email:"cmurphy1337@live.com" String
country:"us" String
dateCreated:2015-01-26 01:07:08.000 Date
verified:true Boolean
banned:false Boolean
lastToken:"" String
lastLogin:2018-09-03 07:55:32.000 Date
lastIP:"127.0.0.1" String
```
For more details on the user object format, see the definition of `UserObject` in main.go

* Alternatively, import the user file from delfruit: ask Klazen for this!
```sql
select id as userid
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

Then reformat to get each object on one line, remove the array wrapper and commas, replace `0/1` booleans with `false/true`, use `{"$date":"2018-10-05 14:48:59Z"}` for dates (https://docs.mongodb.com/compass/master/import-export/)

## How to run

To build and run the project:

```shell
go build && gate-jump.exe
```

Navigate to `http://localhost:10420/`

If you see `{"alive": true}`, you're all set!

Try `http://localhost:10420/user/1` to see your user!