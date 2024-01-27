package main

import (
	"embed"
	_ "embed"
	ui "github.com/malivvan/webkitgtk"
	"io/fs"
	"net/http"
)

//go:embed assets
var assets embed.FS

func main() {
	assets, _ := fs.Sub(assets, "assets")
	app := ui.New(ui.AppOptions{
		ID:   "com.github.malivvan.webkitgtk.examples.handle",
		Name: "WebKitGTK Handle Example",
	})
	app.Handle("main", http.FileServer(http.FS(assets)))
	app.Open(ui.WindowOptions{
		Title:  "Handle Example",
		Width:  200,
		Height: 40,
		URL:    "app://main/",
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
