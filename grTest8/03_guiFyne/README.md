для linux должна быть установлен
sudo apt update
sudo apt install -y libxxf86vm-dev

обязательно компилим мод файл с именем wordcount 
или какаим либо другим... но его указывает как файл запускаы 

cd ~/Project/go/sandbox/grTest8/03_guiFyne
go mod tidy
go build -v -o wordcount ./cmd/wordcount/main.go > build.log 2>&1
ls -l wordcount
cat build.log

тестируем из терминала
./wordcount

с логом ошибок
./wordcount 2> run_error.log
cat run_error.log


ОПЦИОНАЛЬНО с логом билда
cd ~/Project/go/sandbox/grTest8/03_guiFyne
go mod tidy
go build -v -o wordcount ./cmd/wordcount/main.go > build.log 2>&1
ls -l wordcount
cat build.log