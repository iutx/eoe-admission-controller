package main

import "github.com/iutx/eoe-admission-controller/pkg/webhook"

func main() {
	w, err := webhook.New()
	if err != nil {
		panic(err)
	}

	if err = w.Start(); err != nil {
		panic(err)
	}
}
