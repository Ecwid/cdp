package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/ecwid/cdp"
	"github.com/ecwid/cdp/chrome"
)

func main() {
	var err error
	c, err := chrome.New("--window-size=1366,768") // "--headless"
	if err != nil {
		panic(err)
	}
	defer c.Close()

	session, err := c.Client.NewSession(nil)
	if err != nil {
		panic(err)
	}

	wait := func(selector string, appear bool) {
		for start := time.Now(); time.Since(start) < time.Second*20; {
			if appear == (session.Count(selector) > 0) {
				return
			}
			time.Sleep(time.Millisecond * 500)
		}
		panic("wait for " + selector + " failed")
	}

	if err = session.Navigate("https://my.ecwid.com"); err != nil {
		log.Fatal(err)
	}
	wait("input[name='email']", true)
	if err = session.Type("input[name='email']", "XXX"); err != nil {
		log.Fatal(err)
	}
	if err = session.Type("input[name='password']", "XXX"); err != nil {
		log.Fatal(err)
	}
	if err = session.Click("button[id='SIF.sIB']"); err != nil {
		log.Fatal(err)
	}
	wait("button[id='SIF.sIB']", false)

	body, err := session.GetScreenshot(cdp.JPEG, 20, nil, true)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("login.jpeg", body, 0644)

}
