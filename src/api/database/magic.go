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

func (ml *MagicLink) GetMagicLinkFromMagicString() res.ServerError {
	serr := *new(res.ServerError)

	serr.Query = "SELECT * FROM magic WHERE magic = ?"
	serr.Args = append(serr.Args, ml.Magic)
	serr.Err = ml.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	if serr.Err != nil {
		return serr
	}

	return serr
}

func (ml *MagicLink) DeleteMagicLinkFromMagicString() res.ServerError {
	serr := *new(res.ServerError)

	serr.Query = "DELETE FROM magic WHERE magic = ? LIMIT 1"
	serr.Args = append(serr.Args, ml.Magic)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return serr
	}

	return serr
}

// scans all user data into the user struct
func (ml *MagicLink) ScanAll(row *sql.Row) error {

	return row.Scan(
		&ml.ID,
		&ml.UserID,
		&ml.Magic)
}
