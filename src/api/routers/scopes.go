package routers

import (
	"encoding/json"
	"net/http"

	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
)

// register
func createScope(w http.ResponseWriter, r *http.Request) {
	var s database.Scope
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	database.CreateScope(*s.Name, *s.Description)

	res.New(http.StatusOK).JSON(w)
}
