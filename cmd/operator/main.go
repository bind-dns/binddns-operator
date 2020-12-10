package main

import "github.com/bind-dns/binddns-operator/cmd/operator/app"

func main() {
	cmd := app.NewCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
