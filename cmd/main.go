package main

import (
	"github.com/marvinlanhenke/go-distributed-cache/internal/config"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create config")
	}

	app := NewApplication(cfg)

	app.run()
}
