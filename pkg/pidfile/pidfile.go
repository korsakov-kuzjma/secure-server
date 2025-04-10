/*
Пакет pidfile обеспечивает управление PID-файлом
для работы в режиме демона.
*/
package pidfile

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// SavePID сохраняет PID процесса в файл
func SavePID(pidFile string, pid int) error {
	// Проверяем доступ к файлу
	if _, err := os.Stat(pidFile); err == nil {
		return errors.New("PID файл уже существует")
	}

	// Создаем файл
	file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("ошибка создания PID файла: %w", err)
	}
	defer file.Close()

	// Записываем PID
	if _, err := file.WriteString(strconv.Itoa(pid)); err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("ошибка записи PID: %w", err)
	}

	return nil
}

// IsServerRunning проверяет наличие работающего сервера
func IsServerRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// StopProcess останавливает процесс по PID из файла
func StopProcess(pidFile string) error {
	// Читаем PID из файла
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("ошибка чтения PID файла: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("неверный формат PID: %w", err)
	}

	// Находим процесс
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("процесс не найден: %w", err)
	}

	// Отправляем SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("ошибка отправки сигнала: %w", err)
	}

	// Удаляем PID файл
	if err := os.Remove(pidFile); err != nil {
		return fmt.Errorf("ошибка удаления PID файла: %w", err)
	}

	return nil
}
