Сборка исполняемого бинарника
go build -v -o fuelstation ./cmd/fuelstation/main.go > build.log 2>&1

vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./cmd/fuelstation
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/gui
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/db
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/processor
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ 

проверка валидности кода
go vet ./...


отладка при падении Докераdocker

docker-compose down -v --remove-orphans
docker system prune -a --volumes
docker-compose build --no-cache
docker-compose up