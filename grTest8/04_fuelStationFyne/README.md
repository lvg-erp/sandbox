Сборка исполняемого бинарника
go build -v -o fuelstation ./cmd/fuelstation/main.go > build.log 2>&1

vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./cmd/fuelstation
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/gui
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/db
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ go vet ./internal/processor
vladimir@vladimir-VirtualBox:~/Project/go/sandbox/grTest8/04_fuelStationFyne$ 

проверка валидности кода
go vet ./...
