package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	var cfg Config
	err := cleanenv.ReadConfig("config.json", &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	if err := run(cfg); err != nil {
		log.Fatalln(err)
	}
}
