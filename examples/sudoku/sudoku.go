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
		Name:  "sudoku",
		Debug: true,
		Handle: map[string]http.Handler{
			"main": http.FileServer(http.FS(assets)),
		},
	})
	app.Open(ui.WindowOptions{
		Title:  "Sudoku",
		Width:  600,
		Height: 460,
		URL:    "app://main/",
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
