package routers

import (
	"net/http"

	"github.com/badoux/checkmail"
	mux "github.com/labstack/echo"

	log "github.com/spidernest-go/logger"

	//mux "github.com/spidernest-go/mux"
	"golang.org/x/crypto/bcrypt"
)

// validUsername confirms that the given username string is a valid username
func validUsername(username string) bool {
	return len(username) >= 3
}

// validPassword confirms that the given password string is a valid password
func validPassword(password string) bool {
	return len(password) >= 8
}

// createUser creates a user and adds it to database as well as sends a verification email to the provided email
// returns:
//	400 @ invalid username,password,email
//	409 @ user/email already exists
//	500 @ general parsing error
// 	201 @ created user
func createUser(ctx mux.Context) error {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")
	email := ctx.FormValue("email")

	// TODO: Move checkmail functions to mailer package.
	// TODO: checkmail doesn't approve of gmails <email>+<string>@gmail.com format and invalidates it. Consider forking and fixing it.
	// TODO: Confirm that no '.suffix' is required at the end of an email for it to be valid. checkmail says its valid but I am uncertain.
	if !(validUsername(username) && validPassword(password) && (checkmail.ValidateFormat(email) == nil)) {
		return ctx.NoContent(http.StatusBadRequest)
	}

	// TODO: Check database to confirm user/email doesn't already exist.

	// Hash the password.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Err(err).Msgf("Error encrypting password: %v", err)
		return err
	}
	password = string(bytes)

	// TODO: Create the user in the database.

	// TODO: Initalize the magic link for email verification.

	// TODO: Send email to provided email address.

	return ctx.NoContent(http.StatusCreated)
}

// updateUser takes a refresh token, confirms its a valid token, and then updates the user it represents in the user database
// returns:
//	401 @ invalid token provided
//	500 @ ???
// 	204 @ updated successfully
func updateUser(ctx mux.Context) error {
	//token := ctx.FormValue("token")
	//data := ctx.Get("data").(map[string]string) // Format may change depending on database package inputs.

	// TODO: Confirm valid token. (Confirms the user knows who they are updating.)

	// TODO: Pass data into database to update the specified user.

	return ctx.NoContent(http.StatusNoContent)
}

// verifyUser takes a "magic" parameter and crossreferences the user and magic database to confirm correctness and then update their verified status
// returns:
//	400 @ invalid magic link
//	500 @ ???
// 	202 @ verified successfully
func verifyUser(ctx mux.Context) error {
	//magic := ctx.FormValue("magic")

	// TODO: Verify the magic link exists in the database.

	// TODO: Set the verified flag for the user of the given magic link.

	// TODO: Delete the magic link from the database.

	// TODO: Email user to let them know their email is verified.

	return ctx.NoContent(http.StatusAccepted)
}

// deleteUser takes a refresh token, confirms its a valid token, and then flags the user for deactivation after a certain amount of days unless logged in before deactivation
// returns:
//	401 @ invalid token provided
//	500 @ ???
//	202 @ user flagged for deletion
func deleteUser(ctx mux.Context) error {
	//token := ctx.FormValue("token")

	// TODO: Confirm valid token. (Confirms the user knows who they are updating.)

	// TODO: Update users deleted flag in the database.

	return ctx.NoContent(http.StatusAccepted)
}
