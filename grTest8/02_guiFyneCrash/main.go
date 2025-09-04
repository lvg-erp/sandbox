package main

import (
	"fmt"
	"gui/cacheApp"
	"gui/gui"
)

func main() {
	cache := cacheApp.NewCache(3)
	apiGui, err := gui.NewGUIApp(cache)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	apiGui.Run()
}
