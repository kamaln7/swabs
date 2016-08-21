.PHONY: all clean import-dev api-dev
all:
	go build -v ./cmd/importcsv/...
	go build -v ./cmd/api/...
clean:
	rm ./api ./importcsv
import-dev:
	cat /dev/null > db.sql && make && ./importcsv
api-dev:
	make && ./api
