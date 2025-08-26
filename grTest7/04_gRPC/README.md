Создаем файл .proto
Если не установлен proto_buf
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

Для правильной компиляции файл go.mod
должен быть указан так
!!!!module github.com/lvg-erp/sandbox/grTest7/04_gRPC!!!
обязательно запускаем
go mod edit -module github.com/lvg-erp/sandbox/grTest7/04_gRPC
go mod tidy
/// Если не получилось сразу сначала выполняем верхнии команды 
и после
rm -f proto/*.pb.go
protoc --go_out=. --go-grpc_out=. proto/exchange.proto
ls proto/ - это просмотр сгенерированных файлов

генерируем
cd ~/Project/go/sandbox/grTest7/05_grpc_exchange
protoc --go_out=. --go-grpc_out=. proto/exchange.proto

Указываем нужные типы данных

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
После запуска сервиса протестируй его через Postman:

Создай gRPC-запрос:

В Postman: New → gRPC Request.
Укажи localhost:50052 (если изменил порт) или localhost:50051 (если освободил порт).
Импортируй proto/exchange.proto.


Унарный метод (SendMessage):
json{
  "id": "test1",
  "content": "Hello from Postman"
}
Ожидаемый ответ:
json{
  "id": "test1",
  "received": "Processed: Hello from Postman",
  "processed_at": "2025-08-26T13:17:00Z"
}

Стриминг (StreamMessages):
Отправляй сообщения:
json{"id": "stream1", "content": "Stream test 1"}
{"id": "stream2", "content": "Stream test 2"}
Ожидаемый ответ:
json{
  "id": "stream1",
  "received": "Stream processed: Stream test 1",
  "processed_at": "2025-08-26T13:17:00Z"
}
{
  "id": "stream2",
  "received": "Stream processed: Stream test 2",
  "processed_at": "2025-08-26T13:17:00Z"
}

Логи сервера:
bashdocker logs grpc-app
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~