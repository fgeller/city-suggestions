package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type config struct {
	Addr         string
	DataFilePath string
}

func readFlags() (*config, error) {
	var err error
	cfg := new(config)

	flags := flag.NewFlagSet("city-suggestions", flag.ExitOnError)
	flags.StringVar(&cfg.Addr, "addr", "0.0.0.0:7182", "Address to listen to.")
	flags.StringVar(&cfg.DataFilePath, "data-file-path", "data/cities.csv", "CSV containing city data.")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage of city-suggestions:\n\n")
		flags.PrintDefaults()
		help := `
city-suggestions will serve GET HTTP requests on the set address where the query
is an URL parameter named "q" and a location can be passed via the "langitude" and
"longitude" URL parameters. For example, given a GET request like:

GET /suggestions?q=Wok&latitude=43.70011&longitude=-79.4163

It returns a JSON object in the body that looks similar to the following:

{
  "suggestions": [
    {
      "name": "Wokingham",
      "latitude": "51.41120",
      "longitude": "-0.83565",
      "score": 0.9222222222222222
    },
    {
      "name": "Woking",
      "latitude": "51.31903",
      "longitude": "-0.55893",
      "score": 0.4416666666666667
    }
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

	sg, err := newSuggester(cfg.DataFilePath)
	if err != nil {
		log.Fatalf("failed to setup suggester err=%v", err)
	}

	srv := newServer(cfg.Addr, sg)
	err = srv.start()
	if err != nil {
		log.Printf("server start failed err=%v", err)
	}
}
