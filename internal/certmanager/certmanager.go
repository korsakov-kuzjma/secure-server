/*
Пакет certmanager обеспечивает автоматическое получение
и обновление TLS сертификатов от Let's Encrypt.
*/
package certmanager

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"
)

// Manager управляет TLS сертификатами
type Manager struct {
	manager *autocert.Manager
}

// New создает новый менеджер сертификатов
func New(domain, email string) *Manager {
	// Папка для хранения сертификатов
	certDir := "/etc/ssl/kuzjma.ru"
	if err := os.MkdirAll(certDir, 0700); err != nil {
		log.Fatalf("Не удалось создать директорию для сертификатов: %v", err)
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      email, // Ваш email для Let's Encrypt
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(certDir), // Персистентное хранилище
	}

	return &Manager{
		manager: m,
	}
}

// TLSConfig возвращает конфигурацию TLS
func (m *Manager) TLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: m.manager.GetCertificate,
		MinVersion:     tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}

// HTTPHandler возвращает обработчик для HTTP->HTTPS редиректов
func (m *Manager) HTTPHandler(h http.Handler) http.Handler {
	return m.manager.HTTPHandler(h)
}
