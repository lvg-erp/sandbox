Ограниченный буфер с приоритетами
Описание:
Напишите программу, которая обрабатывает задачи двух типов: приоритетные и обычные. 
Задачи поступают в два канала: priorityCh и normalCh. 
Воркеры сначала обрабатывают задачи из priorityCh, и только если он пуст, берут задачи из normalCh. 
Каждая задача — это число, которое нужно удвоить. 
Используйте sync.Mutex для безопасного сохранения результатов 
и sync.WaitGroup для ожидания завершения. 
Если приходит сигнал отмены через контекст, 
воркеры должны завершить обработку и записать ошибку отмены.
Требования:

Вход: два среза чисел priorityTasks и normalTasks, количество воркеров workers.
Результаты: срез []Result, где:
gotype Result struct {
Input  int
Output int
IsPriority bool
}

Ошибки: срез []Error, где:
gotype Error struct {
WorkerID int
Error    error
}

Воркеры должны отдавать предпочтение задачам из priorityCh.
Результаты сортируются по Input.
Используйте контекст для отмены обработки.

Шаблон вызова:
gofunc processTasksWithPriority(ctx context.Context, priorityTasks, normalTasks []int, workers int) ([]Result, []Error, error) 
