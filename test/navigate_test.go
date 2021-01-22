package test

import (
	"context"
	"strings"
	"testing"

	"github.com/ecwid/cdp"
)

func link(t *testing.T, sess *cdp.Session, l string) {
	t.Helper()
	e, err := sess.Query(l)
	check(t, err)
	err = e.Click()
	check(t, err)
}

func checkNav(t *testing.T, sess *cdp.Session, suffix string) {
	t.Helper()
	nav, err := sess.GetNavigationEntry()
	check(t, err)
	if !strings.HasSuffix(nav.URL, suffix) {
		t.Fatalf("%s != %s", suffix, nav.URL)
	}
}

func TestNavigateHistory(t *testing.T) {
	t.Parallel()

	chrome, err := cdp.Launch(context.TODO(), "--disable-popup-blocking", "--headless")
	check(t, err)

	defer chrome.Close()
	sess, err := chrome.Session()
	check(t, err)

	check(t, sess.Navigate(getFilepath("navigate.html")))

	link(t, sess, "[id='a1']")
	link(t, sess, "[id='a2']")
	link(t, sess, "[id='a3']")
	checkNav(t, sess, "#nav3")

	check(t, sess.NavigateHistory(-1))
	checkNav(t, sess, "#nav2")

	check(t, sess.NavigateHistory(-1))
	checkNav(t, sess, "#nav1")

	check(t, sess.NavigateHistory(-1))
	checkNav(t, sess, "navigate.html")

	check(t, sess.NavigateHistory(+1))
	checkNav(t, sess, "#nav1")

	check(t, sess.NavigateHistory(+1))
	checkNav(t, sess, "#nav2")

	check(t, sess.NavigateHistory(+1))
	checkNav(t, sess, "#nav3")

	check(t, sess.NavigateHistory(+1))
	checkNav(t, sess, "#nav3")
}
