package server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/korsakov-kuzjma/secure-server/internal/certmanager"
)

// Server представляет HTTP/HTTPS сервер с поддержкой статических файлов
type Server struct {
	httpServer   *http.Server
	listener     net.Listener
	certManager  *certmanager.Manager
	wg           sync.WaitGroup
	shutdownChan chan struct{}
	mux          *http.ServeMux
	domain       string
	staticDir    string
}

// New создает новый экземпляр сервера
func New(domain, staticDir string) *Server {
	// Инициализация менеджера сертификатов
	certManager := certmanager.New(domain, "korsakov.kuzjma@gmail.com")

	// Создание маршрутизатора
	mux := http.NewServeMux()

	// Конфигурация HTTP сервера
	srv := &http.Server{
		Handler:      mux,
		Addr:         ":443",
		TLSConfig:    certManager.TLSConfig(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	server := &Server{
		httpServer:   srv,
		certManager:  certManager,
		shutdownChan: make(chan struct{}),
		mux:          mux,
		domain:       domain,
		staticDir:    staticDir,
	}

	// Основной обработчик
	mux.HandleFunc("/", server.serveStaticOrFallback)

	return server
}

// serveStaticOrFallback обрабатывает запросы, отдавая статику или переходя к API
func (s *Server) serveStaticOrFallback(w http.ResponseWriter, r *http.Request) {
	// Блокируем попытки path traversal
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Формируем путь к файлу
	path := filepath.Join(s.staticDir, r.URL.Path)
	if r.URL.Path == "/" {
		path = filepath.Join(s.staticDir, "index.html")
	}

	// Проверяем существование файла
	if fileInfo, err := os.Stat(path); err == nil {
		// Запрещаем доступ к директориям
		if fileInfo.IsDir() {
			http.NotFound(w, r)
			return
		}

		// Устанавливаем заголовки кэширования для статики
		if r.URL.Path != "/" {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		http.ServeFile(w, r, path)
		return
	}

	// Если файл не найден и это корень - отдаем 404
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	// Для API-маршрутов
	s.mux.ServeHTTP(w, r)
}

// HandleFunc регистрирует обработчик для API-маршрутов
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

// Start запускает сервер
func (s *Server) Start() error {
	s.wg.Add(1)
	go s.runHTTPServer()

	// Ожидаем сигналы завершения
	select {
	case <-s.shutdownChan:
		log.Println("Получен сигнал завершения")
	case <-s.getInterruptChannel():
		log.Println("Получен сигнал прерывания")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// runHTTPServer запускает HTTP/HTTPS серверы
func (s *Server) runHTTPServer() {
	defer s.wg.Done()

	// HTTP->HTTPS редирект
	go func() {
		log.Printf("Запуск HTTP редиректа на :80 для %s\n", s.domain)
		if err := http.ListenAndServe(":80", s.certManager.HTTPHandler(nil)); err != nil {
			log.Printf("Ошибка HTTP сервера: %v\n", err)
		}
	}()

	// HTTPS сервер
	log.Printf("Запуск HTTPS сервера на :443 для %s\n", s.domain)
	listener, err := tls.Listen("tcp", s.httpServer.Addr, s.httpServer.TLSConfig)
	if err != nil {
		log.Printf("Ошибка создания listener: %v\n", err)
		return
	}
	s.listener = listener

	if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
		log.Printf("Ошибка HTTPS сервера: %v\n", err)
	}
}

// Stop останавливает сервер
func (s *Server) Stop() {
	close(s.shutdownChan)
	s.wg.Wait()
}

// getInterruptChannel возвращает канал для системных сигналов
func (s *Server) getInterruptChannel() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	return c
}
