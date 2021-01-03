package main

import "github.com/bind-dns/binddns-operator/cmd/controller/app"

func main() {
	c := app.NewCommand()
	if err := c.Execute(); err != nil {
		panic(err)
	}
}
