package ui

import "github.com/webview/webview"

// TODO: Сделать GUI

func Create(debug bool) {
	w := webview.New(debug)
	//defer w.Destroy()
	w.SetTitle("Minimal webview example")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")
	w.Run()
}
