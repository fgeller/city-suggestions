package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type config struct {
	Addr string
}

func readFlags() (*config, error) {
	var err error
	cfg := new(config)

	flags := flag.NewFlagSet("city-suggestions", flag.ExitOnError)
	flags.StringVar(&cfg.Addr, "addr", "0.0.0.0:7182", "Address to listen to.")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage of city-suggestions:\n\n")
		flags.PrintDefaults()
		help := `
city-suggestions will serve GET HTTP requests on the set address where the query
is an URL parameter named "q" and a location can be passed via the "langitude" and
"longitude" URL parameters. It returns a JSON objects that looks like the following
example:

{
  "suggestions": [
    {
      "name": "Wokingham",
      "latitude": "51.4112",
      "longitude": "-0.83565",
      "score": 0.8
    },
    {
      "name": "Woking",
      "latitude": "51.31903",
      "longitude": "-0.55893",
      "score": 0.6
    },
  ]
}
`
		fmt.Fprintf(flags.Output(), help)
	}

	err = flags.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func main() {
	cfg, err := readFlags()
	if err != nil {
		log.Fatalf("failed to read flags err=%v", err)
	}

	srv := newServer(cfg.Addr)
	err = srv.start()
	if err != nil {
		log.Printf("server start failed err=%v", err)
	}
}
