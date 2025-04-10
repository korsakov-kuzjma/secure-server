/*
Пакет server реализует основной HTTP/HTTPS сервер
с поддержкой Let's Encrypt и graceful shutdown.
*/
package server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/korsakov-kuzjma/secure-server/internal/certmanager"
)

// Server представляет HTTP/HTTPS сервер
type Server struct {
	httpServer   *http.Server         // Базовый HTTP сервер
	listener     net.Listener         // Сетевой listener
	certManager  *certmanager.Manager // Менеджер сертификатов
	wg           sync.WaitGroup       // Для контроля горутин
	shutdownChan chan struct{}        // Канал для graceful shutdown
	mux          *http.ServeMux       // Маршрутизатор
	domain       string               // Доменное имя
}

// New создает новый экземпляр сервера
func New(domain string) *Server {
	// Инициализация менеджера сертификатов
	certManager := certmanager.New(domain, "korsakov.kuzjma@gmail.com")

	// Создание маршрутизатора
	mux := http.NewServeMux()

	// Конфигурация HTTP сервера
	srv := &http.Server{
		Handler:      mux,
		Addr:         ":443", // HTTPS порт
		TLSConfig:    certManager.TLSConfig(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		httpServer:   srv,
		certManager:  certManager,
		shutdownChan: make(chan struct{}),
		mux:          mux,
		domain:       domain,
	}
}

// HandleFunc регистрирует обработчик для пути
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

// runHTTPServer запускает HTTP/HTTPS серверы в горутинах
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
