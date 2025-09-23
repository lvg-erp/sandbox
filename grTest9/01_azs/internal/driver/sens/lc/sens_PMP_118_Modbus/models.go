package sens_PMP_118_Modbus

type SensPMP118ModbusOtherStatus struct {
	Address         string `json:"Address"`
	OtherParameters map[string]float32
	OtherTables     map[string][]float32
	DriverName      string `json:"DriverName"`
	DevicePNumber   string `json:"DevicePNumber"`
}

type SensPMP118ModbusMainStatus struct {
	Address       string    `json:"Address"`
	H             float32   // Уровень жидкости
	T             float32   // Температура жидкости
	Pr            float32   // Процентное заполнение
	U             float32   // Объем жидкости
	G             float32   // Масса
	R             float32   // Плотность
	U1            float32   // Объем основного продукта
	H2            float32   // Уровень раздела сред
	Ut            float32   // Объем продукта, приведенный к стандартным условиям
	Rt            float32   // Плотность, приведенная к стандартным условиям
	Ri            float32   // Измеренная плотность
	Tr            float32   // Температура измерения плотности
	U2            float32   // Объем жидкости под разделом сред
	Dg            float32   // Абсолютная погрешность измерения массы
	Ts            float32   // Начальная температура
	Nt            []float32 // Показания датчиков температуры
	DriverName    string
	DevicePNumber string
}
