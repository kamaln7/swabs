# swabs.ink

## dev

- `make` - build all binaries
- `make clean` - delete all compiled binaries
- `make importcsv` - compile importcsv
- `make api` - compile api
- `make importcsv-dev` - wipe database, compile, and run ./importcsv
- `make api-dev` - compile and run ./api

## binaries

- `./importcsv` - import data.csv into the SQLite database
- `./api` - run the API
    - environemnt variables:
        - `API_ADDR` - the host:port to listen on
    - api paths:
        - `/v1/brands` - get a JSON array of the brands present in the database
        - `/v1/brands/{brand_name}/inks` - get a JSON array of {brand_name}'s inks
        - `/v1/brands/{brand_name}/inks/{ink_name}` - get a JSON object with info about {brand_name}'s {ink_name}
