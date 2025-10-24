//Доработки на будущее
Случайные задержки:
goimport "math/rand"
time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

Сохранение результатов:
goimport "encoding/json"
results := []int{}
// В processTasks: results = append(results, task.ID)
jsonData, _ := json.Marshal(results)
os.WriteFile("tasks.json", jsonData, 0644)

Ограничение горутин:
gosem := make(chan struct{}, 5)
sem <- struct{}{}
defer func() { <-sem }()