package database

import (
	"database/sql"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
)

type MagicLink struct {
	ID int64 `json:"id"`
	// Read: SERVER
	// Write: Nobody
	UserID int64 `json:"userid"`
	// Read: SERVER
	// Write: SERVER
	Magic string `json:"magic"`
	// Read: USER
	// Write: SERVER
}

func (ml *MagicLink) CreateMagicLink() res.ServerError {
	serr := *new(res.ServerError)
	result := *new(sql.Result)
	err := *new(error)

	serr.Query = "INSERT INTO magic(userid, magic) VALUES(?, ?)"
	serr.Args = append(serr.Args, ml.UserID, ml.Magic)
	result, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return serr
	}

	ml.ID, err = result.LastInsertId()
	if err != nil {
		log.Wtf(err)
	}

	return serr
}
