package fourchan

import (
	"testing"
)

func TestExtractBoardAndThreadId(t *testing.T) {
	tests := []struct {
		url, board, id string
	}{
		{"https://boards.4chan.org/wsg/thread/921167/cats", "wsg", "921167"},
		{"https://boards.4chan.org/wsg/thread/921167", "wsg", "921167"},
		{"https://boards.4chan.org/g/thread/51971506", "g", "51971506"},
		{"https://boards.4chan.org/g/thread/51971506#p51971506", "g", "51971506"},
	}

	for _, test := range tests {
		board, id, err := extractBoardAndThreadId(test.url)
		if err != nil {
			t.Fatal(err)
		}
		if board != test.board || id != test.id {
			t.Fatalf("%s: %s/%s != %s/%s", test.url, board, id, test.board, test.id)
		}
	}
}

func TestExtractBoardAndThreadIdFails(t *testing.T) {
	_, _, err := extractBoardAndThreadId("https://www.google.com")
	if err == nil {
		t.Fatal("err was nil")
	}
}
