#!/bin/bash

# Конфигурация
PROJECT_DIR="$HOME/secure-server"       # Путь к проекту
BINARY_NAME="server"                    # Имя бинарника
SERVICE_NAME="kuzjma-server.service"    # Имя systemd-сервиса
BUILD_CMD="go build -o $BINARY_NAME ./cmd/server" # Команда сборки

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Функция для обработки ошибок
handle_error() {
  echo -e "${RED}Ошибка на шапе: $1${NC}"
  exit 1
}

# Остановка сервиса
echo -e "${YELLOW}Останавливаю сервис $SERVICE_NAME...${NC}"
sudo systemctl stop $SERVICE_NAME || handle_error "Остановка сервиса"

# Сборка проекта
echo -e "${YELLOW}Собираю проект...${NC}"
cd $PROJECT_DIR || handle_error "Переход в директорию проекта"
$BUILD_CMD || handle_error "Сборка проекта"

# Копирование бинарника
echo -e "${YELLOW}Обновляю бинарник...${NC}"
sudo cp $BINARY_NAME /opt/kuzjma-server/ || handle_error "Копирование бинарника"

# Проверка прав на порты
echo -e "${YELLOW}Обновляю права...${NC}"
sudo setcap 'cap_net_bind_service=+ep' /opt/kuzjma-server/$BINARY_NAME || handle_error "Обновление прав"

# Запуск сервиса
echo -e "${YELLOW}Запускаю сервис...${NC}"
sudo systemctl start $SERVICE_NAME || handle_error "Запуск сервиса"

# Проверка статуса
echo -e "\n${GREEN}Статус сервиса:${NC}"
systemctl status $SERVICE_NAME --no-pager

# Проверка портов
echo -e "\n${GREEN}Прослушиваемые порты:${NC}"
ss -tulpn | grep -E ':80|:443'