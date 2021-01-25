package test

import (
	"context"
	"testing"
	"time"

	"github.com/ecwid/cdp"
)

func TestNew5TabOpen(t *testing.T) {
	t.Parallel()

	chrome, err := cdp.Launch(context.TODO(), "--disable-popup-blocking")
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()
	sess, err := chrome.Session()
	if err != nil {
		t.Fatal(err)
	}

	err = sess.Navigate(getFilepath("new_tab.html"))
	if err != nil {
		t.Fatal(err)
	}
	sess.SetTimeout(time.Millisecond * 4000)
	s2, err := sess.OnTargetCreated(func() {
		e, err := sess.Query("#newtabs")
		if err != nil {
			t.Fatal(err)
		}
		err = e.Click()
		if err != nil {
			t.Fatal(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = s2.Activate()
	if err != nil {
		t.Fatal(err)
	}
	tabs, err := sess.GetTargets()
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 5+1+4 {
		t.Fatalf("not 10 tabs but %d", len(tabs))
	}
}
