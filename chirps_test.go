package main

import "testing"

func TestCleanProfaneWords(t *testing.T) {
	inputString := "I want to kerfuffle your Sharbert for a fornax"
	want := "I want to **** your **** for a ****"

	cleaned := cleanProfaneWords(inputString)

	if want != cleaned {
		t.Errorf("returned: %s\nexpected: %s", cleaned, want)
	}
}
