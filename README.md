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

## How to run

To build and run the project:

```shell
go build && gate-jump.exe
```

Navigate to `http://localhost:10420/`

If you see `{"alive": true}`, you're all set!

Try `http://localhost:10420/user/1` to see your user!