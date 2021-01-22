package test

import (
	"context"
	"testing"
	"time"

	"github.com/ecwid/cdp"
)

func Te1stFrameRefresh(t *testing.T) {
	t.Parallel()

	chrome, err := cdp.Launch(context.TODO())
	check(t, err)
	defer chrome.Close()
	sess, err := chrome.Session()
	check(t, err)

	get := func(sel string) *cdp.Element {
		t.Helper()
		el, err := sess.Query(sel)
		check(t, err)
		return el
	}

	check(t, sess.Navigate(getFilepath("frame_playground.html")))
	fid, err := get("#my_frame").GetFrameID()
	check(t, err)
	check(t, sess.SwitchTo(fid))
	finp := get("#frameInput1")
	check(t, finp.Type("123456"))
	check(t, get("#refresh").Click())
	time.Sleep(time.Second * 2)
	if err := finp.Type("654321"); err != cdp.ErrElementDetached {
		t.Fatalf("not expected error: %s", err.Error())
	}
}

func TestFrameRenew(t *testing.T) {
	t.Parallel()

	chrome, err := cdp.Launch(context.TODO())
	check(t, err)
	defer chrome.Close()
	sess, err := chrome.Session()
	check(t, err)

	get := func(sel string) *cdp.Element {
		t.Helper()
		el, err := sess.Query(sel)
		check(t, err)
		return el
	}

	url := getFilepath("frame_playground.html")

	check(t, sess.Navigate(url))
	fid, err := get("#my_frame").GetFrameID()
	check(t, err)
	check(t, get("#button1").Click())

	check(t, sess.SwitchTo(fid))
	time.Sleep(time.Second * 4)

	if _, err := sess.Query("#frameButton1"); err != cdp.ErrFrameDetached {
		t.Fatalf("not expected error: %s", err.Error())
	}
}
