# gate-jump
![Travis CI Build Status](https://img.shields.io/travis/IWannaCommunity/gate-jump/master.svg?&style=flat-square&logo=travis)
![Go Report Card](https://goreportcard.com/badge/github.com/iwannacommunity/gate-jump?style=flat-square)

Central Authentication Service for Delfruit, IWM, and other fangame community services!

## Environment Setup

1. Go Get the repository from the /src/api directory or clone the repository and install libraries manually.
2. Install MariaDB (MySQL may work, but the last time it was tested it failed).
3. Databases aren't created or auto-migrate yet, so you'll have to do `create database gatejump` in MariaDB/MySQL for now.
4. We use [Task](https://taskfile.dev) as the current Build System, install it.
5. The build environment currently requires GNU Date and Git to be present on the system. So if it currently doesn't, fix that before continuing.
6. Install fileb0x via `go get -i github.com/UnnoTed/fileb0x`
7. Install a SMTP Server, for the purposes of testing we will be using a sink called [Inbucket](https://www.inbucket.org/), but you can use whatever one you like.
8. Install [mkcert](https://mkcert.dev) via `go get -i github.com/FiloSottile/mkcert`. We'll need this for having STARTTLS on the SMTP Server. If you aren't using Inbucket or have a proper SMTP Server with STARTTLS, you may skip this.
9. Build the API server by running `task build` in project root.
10. Create a file and appropriately name the following information under `~/config/config.json`

```json
{
    "host":"0.0.0.0",
    "port":"80",
    "sslPort":"443",
    "database":{
        "username":"root",
        "password":"",
        "dsn":"gatejump"
    },
    "https":{
        "certFile":"",
        "keyFile":""
    },
	"mailer":{
		"host":"localhost",
		"port":"2500",
		"user":"gatejump@inbucket",
		"pass":""
	},
	"superuser":{
		"password": "password"
	}
}
```
11. Run `mkcert cert` and rename the resulting files `cert.pem` to `cert.crt` and `cert-key.pem` to `cert.key` and place them with in the root tree next to the Inbucket binary.
12. Run Inbucket with `INBUCKET_SMTP_TLSENABLED=true ./inbucket -netdebug`
13. Run `./api`
14. Verify the service is running by going to `localhost:80/`

## Web Environnment Setup

* Install Node.js from `https://nodejs.org/en/`
* Run in the root directory
```shell
npm install
```
* Run in the root directory
```shell
npm start
```
