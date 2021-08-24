package main

import (
	"testing"
    "github.com/openshift/pagerduty-short-circuiter/cmd/pdcli/login"
)

func TestPdcli(t *testing.T) {

	assertCorrectMessage := func(t testing.TB, actual, expected string) {
		t.Helper()
		if actual != expected {
			t.Errorf("Actual:%s Expected:%s", actual, expected)
		}
	}

    //Tesing all the functions against an empty file.
	t.Run("login test", func(t *testing.T) {
		actual := login.Sample()
		expected := "logged in"
		assertCorrectMessage(t, actual, expected)
	})}
