# City Suggestions

Serves HTTP GET requests to suggest cities within GB based on a query
and location. 


## Installation

The service ships as a single binary which you can download from the
[Releases](https://github.com/fgeller/city-suggestions/releases) section.

Alternatively, you can build it manually or via

```sh
make build
```

## Usage

```
$ ./city-suggestions --help
Usage of city-suggestions:

  -addr string
        Address to listen to. (default "0.0.0.0:7182")
  -data-file-path string
        CSV containing city data. (default "data/cities.csv")

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
```

## Data

The service assumes that the data is a CSV file containing information about cities.
The service expects that the following data is stored at the given index (zero based):

```
1: name of geographical point (utf8) varchar(200)
4: latitude in decimal degrees (wgs84)
5: longitude in decimal degrees (wgs84)
8: ISO-3166 2-letter country code, 2 characters
```

## Tests

Th service includes black box tests that include testing the build process and
starting the binary. You can run them via

```sh
make test
```
