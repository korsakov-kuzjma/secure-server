#!/bin/bash

# Конфигурация
PROJECT_DIR="$HOME/secure-server"       # Путь к проекту
BINARY_NAME="server"                    # Имя бинарника
SERVICE_NAME="kuzjma-server.service"    # Имя systemd-сервиса
STATIC_DIR="/var/www/kuzjma.ru"         # Директория со статикой
BUILD_CMD="go build -o $BINARY_NAME ./cmd/server"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Функция для обработки ошибок
handle_error() {
  echo -e "${RED}Ошибка на шаге: $1${NC}"
  exit 1
}

# 1. Остановка сервиса
echo -e "${YELLOW}[1/6] Останавливаю сервис $SERVICE_NAME...${NC}"
sudo systemctl stop $SERVICE_NAME || handle_error "Остановка сервиса"

# 2. Сборка проекта
echo -e "${YELLOW}[2/6] Собираю проект...${NC}"
cd $PROJECT_DIR || handle_error "Переход в директорию проекта"
$BUILD_CMD || handle_error "Сборка проекта"

# 3. Копирование бинарника
echo -e "${YELLOW}[3/6] Обновляю бинарник...${NC}"
sudo cp $BINARY_NAME /opt/kuzjma-server/ || handle_error "Копирование бинарника"

# 4. Копирование статических файлов
echo -e "${YELLOW}[4/6] Обновляю статические файлы...${NC}"
sudo mkdir -p $STATIC_DIR || handle_error "Создание директории статики"
sudo cp -r static/* $STATIC_DIR/ || echo -e "${YELLOW}Предупреждение: Нет статических файлов для копирования${NC}"
sudo chown -R www-data:www-data $STATIC_DIR || handle_error "Настройка прав статики"

# 5. Настройка прав
echo -e "${YELLOW}[5/6] Настраиваю права...${NC}"
sudo setcap 'cap_net_bind_service=+ep' /opt/kuzjma-server/$BINARY_NAME || handle_error "Настройка прав портов"
sudo chmod 755 /opt/kuzjma-server/$BINARY_NAME || handle_error "Настройка прав бинарника"

# 6. Запуск сервиса
echo -e "${YELLOW}[6/6] Запускаю сервис...${NC}"
sudo systemctl start $SERVICE_NAME || handle_error "Запуск сервиса"

# Проверка статуса
echo -e "\n${GREEN}Статус сервиса:${NC}"
systemctl status $SERVICE_NAME --no-pager

# Проверка портов
echo -e "\n${GREEN}Прослушиваемые порты:${NC}"
ss -tulpn | grep -E ':80|:443'

# Проверка статических файлов
echo -e "\n${GREEN}Проверка статических файлов:${NC}"
ls -la $STATIC_DIR | head -10