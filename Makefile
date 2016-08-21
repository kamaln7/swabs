.PHONY: all clean importcsv-dev api-dev api importcsv
all:
	make importcsv
	make api
clean:
	rm ./api ./importcsv
api:
	go build -v ./cmd/api/...
importcsv:
	go build -v ./cmd/importcsv/...
importcsv-dev:
	cat /dev/null > db.sql && make && ./importcsv
api-dev:
	make api && ./api
