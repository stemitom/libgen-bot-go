package main

import (
	"libgen-bot/internal/libgen"
)

func main() {
	l := libgen.NewLibGenClient()
	l.Search("peace", 2)
	// b, err := bot.NewBot()
	// if err != nil {
	// 	log.Panic(err)
	// }
	// b.Start()
}
