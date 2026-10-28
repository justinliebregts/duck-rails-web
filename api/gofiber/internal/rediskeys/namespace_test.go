package rediskeys

import "testing"

func TestNamespaceValidation(t *testing.T) {
	for _, value := range []string{"", "*", "game/other", "game?", "[ab]", " game", "Game"} {
		if Validate(value) == nil {
			t.Errorf("accepted unsafe namespace %q", value)
		}
	}
	for _, value := range []string{"grc", "game-prod", "game_staging", "game2"} {
		if err := Validate(value); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("REDIS_NAMESPACE", "grc")
	if Key("session/token") != "grc/session/token" {
		t.Fatal("legacy keys changed")
	}
	t.Setenv("REDIS_NAMESPACE", "next-game")
	if Key("session/token") != "next-game/session/token" {
		t.Fatal("namespace captured before config load")
	}
}
