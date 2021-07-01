package main

import (
	"context"
	"log"
	"time"

	"github.com/ecwid/cdp"
	"github.com/ecwid/cdp/pkg/devtool"
)

func main() {
	c, cancel := context.WithTimeout(context.TODO(), time.Minute*30)
	defer cancel()
	browser, err := cdp.Launch(c)
	if err != nil {
		panic(err)
	}
	version, _ := browser.GetVersion()
	log.Printf("%+v", version)

	defer browser.Close()
	sess, err := browser.Session()
	if err != nil {
		panic(err)
	}

	browser.GetWSClient().SetLogLevel(cdp.LevelProtocolFatal)

	sess.SetTimeout(time.Second * 30)
	if err = sess.Navigate("https://mdemo.ecwid.com/", devtool.NetworkIdle); err != nil {
		panic(err)
	}

	all, err := sess.QueryAll(".ec-static-container .grid-product")
	if err != nil {
		panic(err)
	}

	for _, card := range all {
		titleElement, err := card.Query(".grid-product__title-inner")
		if err != nil {
			panic("title is not exist: " + err.Error())
		}
		title, err := titleElement.GetText()
		if err != nil {
			panic("can't read title: " + err.Error())
		}
		priceElement, err := card.Query(".grid-product__price-amount")
		if err != nil {
			panic("price is not exist: " + err.Error())
		}
		price, err := priceElement.GetText()
		if err != nil {
			panic("can't read price: " + err.Error())
		}
		log.Printf("title = %s, price = %s", title, price)
	}
}
