package main

import (
	"os"
	"testing"
)

/*
	This file is mainly used for testing if we can load and start all relevant
	services for starting the server.
	If these don't work than the server can't start.
	Since none of the functions actually return values, all we can do is run them
	to see if an error occurs. The most importnat thing this handles is checking if
	the configuration file.
	Currently this does not test debug mode but the only difference for debug mode should be
	logging to standard out and not to file.
*/

// TestMain tests the Main function for starting the server in debug and non-debug mode.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestCheckDebug confirms if there exists any errors in initalization of logging.
// Errors for this likely occur from ??
func TestSetupLogging(t *testing.T) {
	SetupLogging()
}

// TestBuildInfo confirms if there exists any errors with compiling the build information for the server.
// Errors for this likely occur from failing to `task build` before running the program.
func TestBuildInfo(t *testing.T) {
	BuildInfo()
}

// TestLoadConfiguration confirms if there exisets any errors when loading the configuration data from the json config file.
// Errors for this likely occur from forgetting to create the config.json file under the 'config/' folder.
func TestLoadConfiguration(t *testing.T) {
	LoadConfiguration()
}

// TestStartMailer confirms if there exists any errors when starting the mailing service.
// Errors for this likely occur from failing to start the mailing service before running the server.
func TestStartMailer(t *testing.T) {
	StartMailer()
}

// TestConnectDatabase confirms if there exists any errors when connecting to the database.
// Errors for this likely occur from invalid configuration parameters for logging into the database.
func TestConnectDatabase(t *testing.T) {
	ConnectDatabase()
}

// TestVerifySchemas confirms if there exists any errors when verifying the database schemas.
// Errors for this likely occur from failing to `task build` before running these tests.
func TestVerifySchemas(t *testing.T) {
	VerifySchemas()
}

// TestListenAndServe confirms if the router can properly begin serving api requests.
// Errors for this likely occur from invalid configuration parameters.
func TestListenAndServe(t *testing.T) {
	go ListenAndServe() // goroutine or else it will time out the tests
}
