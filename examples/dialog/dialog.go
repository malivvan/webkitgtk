package main

import (
	ui "github.com/malivvan/webkitgtk"
)

type DialogAPI struct {
	app *ui.App
}

type MessageDialog struct {
	Title   string   `json:"title"`
	Message string   `json:"message"`
	Actions []string `json:"actions"`
}

type OpenDialog struct {
	Title    string `json:"title"`
	Multiple bool   `json:"multiple,omitempty"`
}

type SaveDialog struct {
	Title string `json:"title"`
}

func (a *DialogAPI) Ask(req MessageDialog) (int, error) {
	return <-a.app.Dialog().Ask(req.Title, req.Message, req.Actions...).Show(), nil
}
func (a *DialogAPI) Info(req MessageDialog) (int, error) {
	return <-a.app.Dialog().Info(req.Title, req.Message).Show(), nil
}

func (a *DialogAPI) Warn(req MessageDialog) (int, error) {
	return <-a.app.Dialog().Warn(req.Title, req.Message).Show(), nil
}

func (a *DialogAPI) Fail(req MessageDialog) (int, error) {
	return <-a.app.Dialog().Fail(req.Title, req.Message).Show(), nil
}

func (a *DialogAPI) Open(req OpenDialog) ([]string, error) {
	return <-a.app.Dialog().Open(req.Title).AllowsMultipleSelection(req.Multiple).Show(), nil
}

func (a *DialogAPI) Save(req SaveDialog) (string, error) {
	return <-a.app.Dialog().Save(req.Title).Show(), nil
}

func main() {
	app := ui.New(ui.AppOptions{
		ID:   "com.github.malivvan.webkitgtk.examples.dialog",
		Name: "WebKitGTK Dialog Example",
	})
	app.Open(ui.WindowOptions{
		Title:  "dialog",
		Width:  400,
		Height: 220,
		HTML: `<doctype html>
		<html>
			<head>
				<style>
					input, textarea {
						width: 100%;
					}
					textarea {
						resize: none;
					}
				</style>
			<body>
				<input id="title" type="text" value="title"></input><br>
				<input id="message" type="text" value="message"></input><br>
				<button id="ask">ask</button>
				<button id="info">info</button>
				<button id="warn">warn</button>
				<button id="fail">fail</button>
				<button id="file">open file</button>
				<button id="files">open files</button>
				<button id="save">save</button>
				<br>
				<textarea id="out" rows="8"></textarea>
				<script>
					let title = document.querySelector("#title");
					let message = document.querySelector("#message");
					let output = document.querySelector("#out");
					document.querySelector("#ask").addEventListener("click", () => {
						dialog.ask({title: title.value, message: message.value, actions: ["ok", "cancel"]}).then(resp => output.value = resp);
					});	
					document.querySelector("#info").addEventListener("click", () => {
						dialog.info({title: title.value, message: message.value}).then(resp => output.value = resp);
					});
					document.querySelector("#warn").addEventListener("click", () => {
						dialog.warn({title: title.value, message: message.value}).then(resp => output.value = resp);
					});
					document.querySelector("#fail").addEventListener("click", () => {
						dialog.fail({title: title.value, message: message.value}).then(resp => output.value = resp);
					});
					document.querySelector("#file").addEventListener("click", () => {
						dialog.open({title: title.value}).then(resp => output.value  = resp.join("\n"));
					});
					document.querySelector("#files").addEventListener("click", () => {
						dialog.open({title: title.value, multiple: true}).then(resp => output.value = resp.join("\n"));
					});
					document.querySelector("#save").addEventListener("click", () => {
						dialog.save({title: title.value}).then(resp => output.value = resp);	
					});
				</script>
			</body>
		</html>`,
		Define: map[string]interface{}{
			"dialog": &DialogAPI{app: app},
		},
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
