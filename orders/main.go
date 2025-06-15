package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	cfg := NewDefaultConfig()
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal(err)
	}
	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}
