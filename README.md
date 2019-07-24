# About cdp packages
Golang client driving Chrome browser using the Chrome DevTools Protocol.
CDP Driver has Selenium like interface and tolerance of timing problems so painless can be used for automation of reactive and ajax pages.

packages:
- ecwid/cdp/chrome - to launch Chrome browser
- ecwid/cdp/driver - to painless driving Chrome
- ecwid/cdp - CDP methods to low level interaction

## Installation
`go get -u github.com/ecwid/cdp`

## How to use

Here is an example of using:
```go
package main

import (
	"io/ioutil"
	"log"

	"github.com/ecwid/cdp/chrome"
	"github.com/ecwid/cdp/driver"
)

func main() {
	c, err := chrome.New("--headless", "--window-size=1366,768")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	page := driver.NewDriver(c.Client)

	page.Logger = func(method string, params []interface{}, result []interface{}) {
		log.Printf("%s(%+v) -> %+v", method, params, result)
	}

	page.PanicInterceptor = func(err error) {
		// ignore panic
		log.Printf("err: %s", err.Error())
	}

	page.Navigate("https://my.ecwid.com")
	page.Type("input[name='email']", "XXX")
	page.Type("input[name='password']", "YYY")
	page.Click("button[id='SIF.sIB']")
	page.NotExist("button[id='SIF.sIB']")

	body := page.GetScreenshot(driver.JPEG, 20, true)
	ioutil.WriteFile("login.jpeg", body, 0644)
}

```

Output
```
2019/07/24 15:41:46 Navigate([https://my.ecwid.com]) -> []
2019/07/24 15:41:48 Type([input[name='email'] XXX []]) -> []
2019/07/24 15:41:48 Type([input[name='password'] YYY []]) -> []
2019/07/24 15:41:48 Click([button[id='SIF.sIB']]) -> []
2019/07/24 15:42:48 Predicate([button[id='SIF.sIB'] have to disappear]) -> []
2019/07/24 15:42:48 err: sessionID: 87E08ED8339D9CF72CEBB359A14A3F82
error: expected condition `button[id='SIF.sIB'] have to disappear` not satisfied
```

Replace XXX to your ecwid email and YYY to password to get code working.