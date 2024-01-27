package main

import (
	_ "embed"
	ui "github.com/malivvan/webkitgtk"
)

type NotifyAPI struct {
	app *ui.App
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (a *NotifyAPI) Notify(notification *Notification) (uint32, error) {
	return a.app.Notify(notification.Title, notification.Body).
		Action("open", func() {
			println("action open")
		}).
		Closed(func() {
			println("closed")
		}).Show()

}

func main() {

	app := ui.New(ui.AppOptions{
		ID:   "com.github.malivvan.webkitgtk.examples.notify",
		Name: "WebKitGTK Notify Example",
	})

	app.Open(ui.WindowOptions{
		Title:  "notify",
		Width:  200,
		Height: 90,
		HTML: `<doctype html>
		<html>
			<body>
				<input id="title" type="text" placeholder="Title"></input><br>
				<input id="body" type="text" placeholder="Body"></input><br>
				<button id="notify">notify</button>
				<script>
					document.querySelector("#notify").addEventListener("click", () => {
						let title = document.querySelector("#title");
						let body = document.querySelector("#body");
						api.notify({
							title: title.value,
							body: body.value,	
						}).catch(err => console.error("error: " + err));
					});	
				</script>
			</body>
		</html>`,
		Define: map[string]interface{}{
			"api": &NotifyAPI{app: app},
		},
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
