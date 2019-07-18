# WIP cdp
Golang client driving Chrome browser using the Chrome DevTools Protocol.
CDP client has selenium like interface and can be used for any kind of browser automation.


## Installation
`go get -u github.com/ecwid/cdp`

## How to use

Here is an example of using `cdp`:
```go
package main

import (
	"github.com/ecwid/cdp/chrome"
)

func main() {
	c, err := chrome.New("--headless", "--window-size=1366,768")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	driver := chrome.NewDriver(c.Client)

	driver.Navigate("https://my.ecwid.com")
	driver.Type("input[name='email']", "XXX")
	driver.Type("input[name='password']", "YYY")
	driver.Click("button[id='SIF.sIB']")
	driver.NotExist("button[id='SIF.sIB']")
  ...
}
```

chrome.Driver just wrapper under CDP client's session used to make automation more stable and reliable.
