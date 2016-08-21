# swabs.ink

## dev

- `make` - build all binaries
- `make clean` - delete all compiled binaries
- `make import-dev` - wipe database, compile, and ./importcsv
- `make api-dev` - compile and ./api

## binaries

- `./importcsv` - import data.csv into the SQLite database
- `./api` - run the API
    - environemnt variables:
        - `API_ADDR` - the host:port to listen on
    - api paths:
        - `/v1/brands` - get a JSON array of the brands present in the database

