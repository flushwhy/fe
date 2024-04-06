package cmd_test

import (
	"testing"

	"codeberg.org/flush/fe/cmd"
)

func TestButlerPusher(t *testing.T) {
	// Test case when the directory can be read successfully
	directory := "testdata/cmd/butler_pusher"
	username := "flush"
	game := "xo"
	userversion := ""
	t.Run("Directory Read Success", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when the subdirectory can be read successfully
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = ""
	t.Run("Subdirectory Read Success", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when user version is empty
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = ""
	t.Run("UserVersion Empty", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})

	// Test case when user version is not empty
	directory = "testdata/cmd/butler_pusher"
	username = "ff"
	game = "xo"
	userversion = "1.0.0"
	t.Run("UserVersion Not Empty", func(t *testing.T) {
		testButlerPusher(t, directory, username, game, userversion)
	})
}

func testButlerPusher(t *testing.T, directory, username, game, userversion string) {
	cmd.Butler_pusher(username, game, directory, userversion)
}
