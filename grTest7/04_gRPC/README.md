Создаем файл .proto
Если не установлен proto_buf
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
генерируем
cd ~/Project/go/sandbox/grTest7/05_grpc_exchange
protoc --go_out=. --go-grpc_out=. proto/exchange.proto

Указываем нужные типы данных
Для правильной компиляции файл go.mod
должен быть указан так
!!!!module github.com/lvg-erp/sandbox/grTest7/04_gRPC!!!
обязательно запускаем
go mod edit -module github.com/lvg-erp/sandbox/grTest7/04_gRPC
go mod tidy
