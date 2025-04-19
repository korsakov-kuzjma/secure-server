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
	pidFile   = "/var/run/kuzjma-server/kuzjma-server.pid"
	staticDir = "/var/www/kuzjma.ru" // Директория со статическими файлами
)

func main() {
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

	startServer()
}

func startServer() {
	if pidfile.IsServerRunning(pidFile) {
		fmt.Println("Сервер уже запущен")
		os.Exit(1)
	}

	if err := pidfile.SavePID(pidFile, os.Getpid()); err != nil {
		log.Fatalf("Ошибка сохранения PID: %v\n", err)
	}

	// Создаем экземпляр сервера с указанием домена и директории статики
	srv := server.New("kuzjma.ru", staticDir)

	// Регистрируем API-маршруты
	srv.HandleFunc("/api/status", apiStatusHandler)
	srv.HandleFunc("/api/data", apiDataHandler)

	log.Println("Запуск сервера для kuzjma.ru...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Ошибка сервера: %v\n", err)
	}
}

func stopServer() {
	if err := pidfile.StopProcess(pidFile); err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}
	fmt.Println("Сервер успешно остановлен")
}

func checkStatus() {
	if pidfile.IsServerRunning(pidFile) {
		fmt.Println("Сервер работает")
	} else {
		fmt.Println("Сервер не запущен")
	}
}

func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  server start    - запустить сервер")
	fmt.Println("  server stop     - остановить сервер")
	fmt.Println("  server status   - проверить статус")
	fmt.Println("  server          - запустить сервер (без демонизации)")
}

// Пример обработчика API
func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok", "service": "kuzjma.ru"}`)
}

func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	// Ваша логика обработки API
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": [1, 2, 3]}`)
}
