# webkitgtk [![Go Reference](https://pkg.go.dev/badge/github.com/malivvan/webkitgtk.svg)](https://pkg.go.dev/github.com/malivvan/webkitgtk) [![Release](https://img.shields.io/github/v/release/malivvan/webkitgtk.svg?sort=semver)](https://github.com/malivvan/webkitgtk/releases/latest) [![Go Report Card](https://goreportcard.com/badge/github.com/malivvan/webkitgtk)](https://goreportcard.com/report/github.com/malivvan/webkitgtk) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
Pure Go WebKitGTK binding for **Linux** and **FreeBSD**.

> This is pre release software so expect bugs and potentially API breaking changes
> but each release will be tagged to avoid breaking people's code.

## Installation

```sh
# go 1.21.5+
go get github.com/malivvan/webkitgtk@latest
```

## Example
The following example shows how to create a simple GTK window with a button that closes the application when clicked.
```go
package main

import ui "github.com/malivvan/webkitgtk"

type API struct {
	app *ui.App
}

func (a *API) Quit() error {
	a.app.Quit()
	return nil
}

func main() {
	app := ui.New(ui.AppOptions{
		Name: "example",
	})
	app.Open(ui.WindowOptions{
		Title: "example",
		HTML:  `<button onclick="app.quit()">quit</button>`,
		Define: map[string]interface{}{
			"app": &API{app: app},
		},
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
```
## Running / Building

Running / building defaults to debug mode outputting logs to stderr. To build in release mode use the `release` build tag.

```sh
go build -tags release -ldflags "-s -w" -trimpath
```

The resulting release binary will be about ~6MB in size and cam be compressed further with [UPX](https://upx.github.io/) to about ~2.5MB.


## Examples

- [echo](examples/echo/echo.go) - call go functions from javascript 
- [dialog](examples/dialog/dialog.go) - application spawning different types of dialog windows
- [handle](examples/handle/handle.go) - handle requests on the app:// uri scheme to serve embedded files
- [notify](examples/notify/notify.go) - application sending different types of notifications
- [systray](examples/systray/systray.go) - example application showing how to use the systray

## Dependencies
Either
[`webkit2gtk-4.1`](https://pkgs.org/search/?q=webkit2gtk-4.1&on=name)
([*stable*](https://webkitgtk.org/reference/webkit2gtk/stable/)) or
[`webkitgtk-6.0`](https://pkgs.org/search/?q=webkitgtk-6.0&on=name)
([*unstable*](https://webkitgtk.org/reference/webkitgtk/unstable/index.html))
is required at runtime. If both are installed the stable version will be used.

<table>
  <tr>
    <td style="font-size: 14px;font-weight: bold;">Debian / Ubuntu</td>
    <td><code>apt install libwebkit2gtk-4.1</code></td>
    <td><code>apt install libwebkitgtk-6.0</code></td>
  </tr>
    <tr>
        <td style="font-size: 14px;font-weight: bold;">RHEL / Fedora</td>
        <td><code>dnf install webkitgtk4</code></td>
        <td><code>dnf install webkitgtk3</code></td>
    </tr>
    <tr>
        <td style="font-size: 14px;font-weight: bold;">Arch</td>
        <td><code>pacman -S webkit2gtk-4.1</code></td>
        <td><code>pacman -S webkitgtk-6.0</code></td>
    </tr>
    <tr>
        <td style="font-size: 14px;font-weight: bold;">Alpine</td>
        <td><code>apk add webkit2gtk</code></td>
        <td><code>apk add webkit2gtk-6.0</code></td>
    </tr>
    <tr>
        <td style="font-size: 14px;font-weight: bold;">Gentoo</td>
        <td colspan="2" align="center"><code style="margin:0px;padding:2px">emerge -av net-libs/webkit-gtk</code></td>
    </tr>
    <tr>
        <td style="font-size: 14px;font-weight: bold;">FreeBSD</td>
        <td><code>pkg install webkit2-gtk3</code></td>
        <td><code>pkg install webkit2-gtk4</code></td>
    </tr>
</table>

## License
This project is licensed under the [MIT License](LICENSE).