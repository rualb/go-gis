# go-gis

`go-gis` is a microservice focused on Geographic Information System (GIS) capabilities, spatial data, and mapping functionalities.

## Features

- **Spatial Data:** Handles geographical and location-based data.
- **Database:** Uses GORM with PostgreSQL (often coupled with PostGIS in the deployment environment).
- **HTTP Server:** REST API powered by the Echo framework.
- **Metrics:** Prometheus metrics enabled.

## Prerequisites

- Go 1.26+
- Python 3.x

## Build and Run

```sh
# Run tests
python Makefile.py test

# Build binary for Linux
python Makefile.py linux
```

## Architecture Context

This service processes all location-related queries and provides map data functionalities. It interacts heavily with the central database and is orchestrated via the API gateway (`go-proxy`).
