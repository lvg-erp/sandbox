Напишите программу на Go, которая реализует пул воркеров для обработки задач из API. 
Каждая задача — это запрос к API для получения данных о сотрудниках (Employee), их обработка (например, вычисление бонуса на основе стажа) 
и сохранение результата в общий срез результатов. Программа должна:

Использовать конкурентность (горутины и каналы).
Обрабатывать ошибки (например, если API возвращает ошибку).
Прерывать выполнение при первой ошибке с использованием контекста.
Иметь тест для проверки корректности работы.

Эта задача объединяет темы конкурентности, работы с API и тестирования, которые ты уже изучал, 
и добавляет реалистичность, так как такие задачи часто встречаются на собеседованиях

Примерный план решения 
1. **Структуры**:
    - `Employee`: Структура для данных сотрудника (ID, имя, стаж, бонус, флаг обработки).
    - `Result`: Структура для результата обработки (сотрудник или ошибка).
    - `MockAPI`: Имитация API с методом `GetEmployee`, который возвращает данные сотрудника или ошибку.

2. **Пул воркеров**:
    - Создаётся канал `taskCh` для отправки ID сотрудников и канал `resultCh` для результатов.
    - Запускается `workers` горутин, каждая из которых:
        - Читает ID из `taskCh`.
        - Запрашивает данные сотрудника через `api.GetEmployee`.
        - При ошибке отправляет сообщение в `resultCh` и отменяет контекст через `once.Do(cancel)`.
        - При успехе вычисляет бонус и отправляет результат в `resultCh`.

3. **Обработка результатов**:
    - Главный цикл читает из `resultCh`.
    - Если встречается ошибка, устанавливается `hasError = true`, и дальнейшие успешные результаты игнорируются.
    - Результаты сохраняются в срез `results` с использованием `sync.Mutex` для безопасного доступа.

4. **Контекст**:
    - Используется `context.WithCancel` для прерывания обработки при первой ошибке.
    - Воркеры проверяют `ctx.Done()` для остановки при отмене

первое решение
type Employee struct {
ID        int     `json:"id"`
Name      string  `json:"name"`
Years     int     `json:"years"`     // стаж
Bonus     float64 `json:"bonus"`     // премия
Processed bool    `json:"processed"` // флаг обработки
}

type ErrorsInStream struct {
WorkerID int
Error    error
}

// MockAPI имитирует API для получения данных сотрудников
type MockAPI struct{}

// func (api *MockAPI) GetEmployee(ctx context.Context, id int) (Employee, ErrorsInStream, error) {
func (api *MockAPI) GetEmployee(ctx context.Context, id int) (Employee, ErrorsInStream) {
select {
case <-ctx.Done():
return Employee{}, ErrorsInStream{}
default:
time.Sleep(100 * time.Millisecond)
// для примера смоделируем ошибку на ИД 3
if id == 3 {
return Employee{}, ErrorsInStream{
WorkerID: id,
Error:    errors.New("employee not found"),
}
}
return Employee{
ID:    id,
Name:  fmt.Sprintf("Employee %d", id),
Years: id * 2,
}, ErrorsInStream{}
}
}

// Сохраняем результа
type Result struct {
Employee Employee `json:"employee"`
Error    error    `json:"error"`
}

const workers = 3 // число потоков обработки

func main() {
api := &MockAPI{}
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

	taskCh := make(chan int, 10)                //канал по задачам
	resultCh := make(chan Result, 10)           //канал по результатам
	resultErrorsCh := make(chan ErrorsInStream) //канал по ошибам
	var results []Employee
	var resultErrors []ErrorsInStream
	var wg sync.WaitGroup
	var mu sync.Mutex
	//var once sync.Once

	employeesIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, id := range employeesIDs {
		taskCh <- id
	}
	close(taskCh)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for id := range taskCh {
				select {
				case <-ctx.Done():
					resultCh <- Result{Error: fmt.Errorf("worker %d cancelled: %v", workerID, ctx.Err())}
					return
				default:
					employee, err := api.GetEmployee(ctx, id)
					if err.Error != nil {
						//resultCh <- Result{Error: fmt.Errorf("worker %d: error for employee %d: %v", workerID, id, err)}
						resultErrorsCh <- ErrorsInStream{
							WorkerID: id,
							Error:    fmt.Errorf("worker %d: error for employee %d: %v", workerID, id, err),
						}
						//once.Do(cancel)
						return
					}
					//расчет бонусов
					employee.Bonus = float64(employee.Years) * 1000
					employee.Processed = true
					resultCh <- Result{Employee: employee}
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	//читаем из канала результатов
	var hasError bool
	for result := range resultCh {
		//if result.Error != nil {
		//	fmt.Println("Ошибка:", result.Error)
		//	hasError = true
		//
		//	continue
		//}
		mu.Lock()
		results = append(results, result.Employee)
		mu.Unlock()
		fmt.Printf("Результат обработан сотрудник %s, бонус: %.2f\n", result.Employee.Name, result.Employee.Bonus)
		//if !hasError {
		//	mu.Lock()
		//	results = append(results, result.Employee)
		//	mu.Unlock()
		//	fmt.Printf("Результат обработан сотрудник %s, бонус: %.2f\n", result.Employee.Name, result.Employee.Bonus)
		//}
	}

	if len(resultErrorsCh) > 0 {
		hasError = true
		for errResult := range resultErrorsCh {
			mu.Lock()
			resultErrors = append(resultErrors, errResult)
			mu.Unlock()
		}
	}

	
	
	if hasError {
		//fmt.Println("Обработка прервана из-за ошибки, результаты: ", results)
		fmt.Println("Во время обработки были ошибки: ", resultErrors)
	} else {
		fmt.Println("Все сотрудники обработаны, результаты: ", results)
	}

}