package ostep_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChapterByNum(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeHomePage)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ch, err := c.ChapterByNum(context.Background(), 26)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Chapter != 26 {
		t.Errorf("Chapter = %d, want 26", ch.Chapter)
	}
	if ch.Title != "Concurrency and Threads" {
		t.Errorf("Title = %q", ch.Title)
	}
}

func TestChapterByNumNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeHomePage)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.ChapterByNum(context.Background(), 999)
	if err == nil {
		t.Error("expected error for non-existent chapter")
	}
}

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeHomePage)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	results, err := c.Search(context.Background(), "intro")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].Title != "Introduction" {
		t.Errorf("Title = %q", results[0].Title)
	}
}

func TestInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeHomePage)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	info, err := c.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Chapters != 5 {
		t.Errorf("Chapters = %d, want 5", info.Chapters)
	}
	if info.Parts == 0 {
		t.Error("Parts should not be 0")
	}
}
