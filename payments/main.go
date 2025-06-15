package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	cfg := NewDefaultConfig()
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	if err := run(cfg); err != nil {
		log.Fatalln(err)
	}
}
