package main

import (
	_ "embed"
	ui "github.com/malivvan/webkitgtk"
)

//go:embed icon.png
var icon []byte

func main() {
	app := ui.New(ui.AppOptions{
		ID:   "com.github.malivvan.webkitgtk.examples.systray",
		Name: "WebKitGTK Systray Example",
		Hold: true,
	})

	menu := app.Menu(icon)
	menu.Add("open").OnClick(func(checked bool) {
		app.Open(ui.WindowOptions{
			Title:  "tray",
			Width:  160,
			Height: 60,
			HTML: `<doctype html>
		<html>
			<body>
				Tray Example
			</body>
		</html>`,
		})
	})
	submenu := menu.AddSubmenu("submenu")
	radio1 := submenu.AddRadio("radio 1", true).SetIcon(icon).OnClick(func(checked bool) {
		println("radio 1", checked)
	})
	radio2 := submenu.AddRadio("radio 2", false).SetIcon(icon).OnClick(func(checked bool) {
		println("radio 2", checked)
	})
	menu.AddSeparator()
	menu.AddCheckbox("enable submenu items", true).SetIcon(icon).OnClick(func(checked bool) {
		println("enable submenu items", checked)
		radio1.SetDisabled(!checked)
		radio2.SetDisabled(!checked)
	})
	menu.AddCheckbox("show submenu", true).SetIcon(icon).OnClick(func(checked bool) {
		println("show submenu", checked)
		submenu.Item().SetHidden(!checked)
	})
	menu.AddSeparator()
	menu.Add("quit").OnClick(func(checked bool) {
		println("quit")
		app.Quit()
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
