package main

import (
	"log/slog"
	"os"
)

func main() {

	configuration := configuration{
		address:        ":8080",
		databaseConfig: databaseConfig{},
	}

	api := application{
		configuration: configuration,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if error := api.run(api.mount()); error != nil {
		slog.Error("Error starting server", "error", error)
		os.Exit(1)
	}
}
