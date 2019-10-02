# Non maintained!
Using https://github.com/Ecwid/witness instead

# About cdp packages
Golang client driving Chrome browser using the Chrome DevTools Protocol.
CDP Agent has Selenium like interface and tolerance of timing problems so painless can be used for automation of reactive and ajax pages.

packages:
- ecwid/cdp/chrome - to launch Chrome browser
- ecwid/cdp/agent - to painless driving Chrome
- ecwid/cdp - CDP methods to low level interaction

## Installation
`go get -u github.com/ecwid/cdp`

## How to use

Here is an example of using:
```go
package main

import (
	"io/ioutil"

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
	a.Navigate("https://my.ecwid.com")
	a.Type("input[name='email']", "XXX")
	a.Type("input[name='password']", "XXX")
	a.Click("button[id='SIF.sIB']")
	a.NotExist("button[id='SIF.sIB']")

	body := a.GetScreenshot(cdp.JPEG, 20, true)
	ioutil.WriteFile("login.jpeg", body, 0644)
}
```

See https://github.com/Ecwid/cdp/tree/master/examples for more examples
