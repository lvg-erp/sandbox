package web

import "embed"

//go:embed static/*
var staticFiles embed.FS

// GetStaticFiles возвращает встроенные статические файлы
func GetStaticFiles() embed.FS {
	return staticFiles
}

// GetTemplate читает HTML шаблон из встроенных файлов
func GetTemplate() ([]byte, error) {
	return staticFiles.ReadFile("static/index.html")
}
