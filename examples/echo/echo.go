package main

import (
	ui "github.com/malivvan/webkitgtk"
)

type API struct{}

func (a *API) Echo(msg string) (string, error) {
	return msg, nil
}

func main() {
	app := ui.New(ui.AppOptions{
		Name:  "echo",
		Debug: true,
	})
	app.Open(ui.WindowOptions{
		Title:  "echo",
		Width:  420,
		Height: 44,
		HTML: `<doctype html>
		<html>
			<body>
				<input id="in" type="text"></input>
				<button id="echo">echo</button>
				<input id="out" type="text" disabled></input>
				<script>
					document.querySelector("#echo").addEventListener("click", () => {
						let input = document.querySelector("#in");
						let output = document.querySelector("#out");
						api.echo(input.value)
							.then(resp => output.value = resp)
							.catch(err => output.value = "error: " + err);
					});	
				</script>
			</body>
		</html>`,
		Define: map[string]interface{}{
			"appID": "echo",
			"api":   &API{},
		},
	})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
