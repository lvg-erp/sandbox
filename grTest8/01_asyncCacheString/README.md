Описание:
Реализуйте асинхронный кэш, который подсчитывает количество слов в строке (слово — это последовательность символов, разделённых пробелами). 
Если строка уже была обработана, возвращается результат из кэша. 
Если строка новая, воркер выполняет подсчёт слов и сохраняет результат. 
Используйте sync.RWMutex для безопасного доступа к кэшу, каналы для отправки запросов и контекст для отмены. 
Обрабатывайте ошибки, например, если строка пустая.
Требования:

Вход: Срез строк []string для обработки.
Кэш: map[string]int (строка → количество слов).
Результаты: Срез []Result, где:
gotype Result struct {
Input  string
Count  int
}

Ошибки: Срез []Error, где:
gotype Error struct {
Input string
Error error
}

Если строка пустая, возвращать ошибку fmt.Errorf("empty string").
Используйте sync.RWMutex для конкурентного доступа к кэшу.
Результаты сортируются по Input (лексикографически).
Поддержка нескольких воркеров.

Шаблон вызова:
gofunc processStringsWithCache(ctx context.Context, inputs []string, workers int) ([]Result, []Error, error) {
// Реализовать
}
Пример:

Вход: ["hello world", "", "test", "hello world"]
Выход:

Results: [{Input:"hello world" Count:2}, {Input:"test" Count:1}]
Errors: [{Input:"" Error:empty string}]
