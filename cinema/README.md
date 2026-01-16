go install github.com/pressly/goose/v3/cmd/goose@latest
///
Сначала собери образ:
Bashdocker compose build
Затем запусти миграции:
Bashdocker compose run --rm cinema make up
