package test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/ecwid/cdp"
)

func TestClickHit(t *testing.T) {
	t.Parallel()

	var expectedRate int64 = 60 // 60% click hit

	chrome, err := cdp.Launch(context.TODO())
	check(t, err)
	defer chrome.Close()
	s, err := chrome.Session()

	get := func(sel string) *cdp.Element {
		t.Helper()
		el, err := s.Query(sel)
		check(t, err)
		return el
	}

	check(t, err)
	check(t, s.Navigate(getFilepath("click_playground.html")))

	target := get("#target")

	var pass int64
	var miss int64
	for i := 0; i < 50; i++ {
		err := target.Click()
		switch err {
		case nil:
			pass++
		case cdp.ErrElementMissClick:
			miss++
		default:
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 300)
	}

	clickedText, err := target.GetText()
	check(t, err)

	clicked, err := strconv.ParseInt(clickedText, 10, 64)
	check(t, err)

	rate := (100 * pass) / (miss + pass)
	t.Logf("pass = %d, miss = %d, rate = %d", pass, miss, rate)
	if rate <= expectedRate {
		t.Fatalf("miss click degradation - expected at least %d%% success click, but was %d", expectedRate, rate)
	}
	if clicked != pass {
		t.Fatalf("%d flaky clicks", pass-clicked)
	}

}
