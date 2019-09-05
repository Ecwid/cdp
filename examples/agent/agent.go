package main

import (
	"io/ioutil"
	"time"

	"github.com/ecwid/cdp"
	"github.com/ecwid/cdp/agent"
	"github.com/ecwid/cdp/chrome"
)

func main() {
	var err error
	c, err := chrome.New("--window-size=1366,768") // "--headless"
	if err != nil {
		panic(err)
	}
	defer c.Close()

	a := agent.New(c.Client)
	a.Deadline = time.Second * 20

	a.Navigate("https://my.ecwid.com")
	a.Type("input[name='email']", "XXX")
	a.Type("input[name='password']", "XXX")
	a.Click("button[id='SIF.sIB']")
	a.NotExist("button[id='SIF.sIB']")

	body := a.GetScreenshot(cdp.JPEG, 20, true)
	ioutil.WriteFile("login.jpeg", body, 0644)
}
