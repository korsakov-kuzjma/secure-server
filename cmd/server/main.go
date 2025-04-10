/*
Основной исполняемый файл сервера.
Обрабатывает команды start/stop и управляет жизненным циклом сервера.
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/korsakov-kuzjma/secure-server/internal/server"
	"github.com/korsakov-kuzjma/secure-server/pkg/pidfile"
)

const (

	//pidFile = "/var/run/kuzjma-server.pid" // Стандартное расположение PID файла для демонов
	//pidFile = "/var/run/kuzjma-server/kuzjma-server.pid" // Новый путь
	pidFile = "/run/kuzjma-server/kuzjma-server.pid" // Актуальный путь для Ubuntu
)

func main() {
	// Проверяем аргументы командной строки
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "start":
			startServer()
		case "stop":
			stopServer()
		case "status":
			checkStatus()
		default:
			printUsage()
		}
		return
	}

	// Если аргументов нет - запускаем сервер
	startServer()
}

// startServer запускает HTTPS сервер
func startServer() {
	// Проверяем, не запущен ли уже сервер
	if pidfile.IsServerRunning(pidFile) {
		fmt.Println("Сервер уже запущен")
		os.Exit(1)
	}

	// Сохраняем PID текущего процесса
	if err := pidfile.SavePID(pidFile, os.Getpid()); err != nil {
		log.Fatalf("Ошибка сохранения PID: %v\n", err)
	}

	// Создаем экземпляр сервера для домена kuzjma.ru
	srv := server.New("kuzjma.ru")

	// Регистрируем обработчики маршрутов
	srv.HandleFunc("/", rootHandler)
	srv.HandleFunc("/health", healthHandler)

	log.Println("Запуск сервера для kuzjma.ru...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Ошибка сервера: %v\n", err)
	}
}

// stopServer останавливает работающий сервер
func stopServer() {
	if err := pidfile.StopProcess(pidFile); err != nil {
		fmt.Printf("Ошибка остановки: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Сервер успешно остановлен")
}

// checkStatus проверяет статус сервера
func checkStatus() {
	if pidfile.IsServerRunning(pidFile) {
		fmt.Println("Сервер работает")
	} else {
		fmt.Println("Сервер не запущен")
	}
}

// printUsage выводит справку по использованию
func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  server start    - запустить сервер")
	fmt.Println("  server stop     - остановить сервер")
	fmt.Println("  server status   - проверить статус")
	fmt.Println("  server          - запустить сервер (без демонизации)")
}

// rootHandler обрабатывает запросы к корневому пути
func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Добро пожаловать на kuzjma.ru</h1>")
	fmt.Fprintf(w, "<p>Безопасный HTTPS сервер</p>")
	fmt.Fprintf(w, "<p>:)</p>")
}

// healthHandler для проверки здоровья сервера
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
