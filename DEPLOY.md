# 🚀 Go Quiz — Deployment Guide

**Полное руководство по деплою Go Quiz на Ubuntu сервер с Docker**

---

## 📋 Содержание

1. [Требования](#требования)
2. [Быстрый старт](#быстрый-старт)
3. [Настройка сервера](#настройка-сервера)
4. [Docker деплой](#docker-деплой)
5. [SSL/TLS настройка](#ssltls-настройка)
6. [Деплой скриптом](#деплой-скриптом)
7. [Мониторинг](#мониторинг)
8. [Бэкапы](#бэкапы)
9. [Troubleshooting](#troubleshooting)

---

## 📦 Требования

### Минимальные:
- **CPU:** 1 ядро (2+ GHz)
- **RAM:** 1 ГБ
- **Storage:** 10 ГБ
- **OS:** Ubuntu 24.04 LTS

### Рекомендуемые:
- **CPU:** 2 ядра (3+ GHz)
- **RAM:** 2 ГБ
- **Storage:** 20 ГБ SSD
- **OS:** Ubuntu 24.04 LTS

---

## ⚡ Быстрый старт

### 1. Клонирование репозитория
```bash
git clone https://github.com/Folombas/qwen_test.git
cd qwen_test
```

### 2. Настройка окружения
```bash
cp .env.example .env
nano .env  # Измените JWT_SECRET и другие переменные
```

### 3. Запуск Docker
```bash
docker-compose up -d
```

### 4. Проверка
```bash
docker-compose ps
curl http://localhost:8080/api/stats
```

---

## 🖥️ Настройка сервера (Ubuntu 24.04)

### 1. Обновление системы
```bash
sudo apt update && sudo apt upgrade -y
```

### 2. Установка Docker
```bash
# Добавляем репозиторий Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Устанавливаем Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# Добавляем пользователя в группу docker
sudo usermod -aG docker $USER
```

### 3. Установка Docker Compose
```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Проверка
docker-compose --version
```

### 4. Настройка фаервола
```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

---

## 🐳 Docker деплой

### 1. Подготовка
```bash
cd /opt/qwen_test

# Копируем .env
cp .env.example .env
nano .env
```

### 2. Изменение .env
```bash
# Обязательно измените:
JWT_SECRET=ваш-секретный-ключ-минимум-32-символа
DOMAIN_NAME=ваш-домен.com
SSL_EMAIL=ваш@email.com
```

### 3. Запуск
```bash
# Build и запуск
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f app
```

### 4. Остановка
```bash
docker-compose down
```

### 5. Перезапуск
```bash
docker-compose restart
```

---

## 🔒 SSL/TLS настройка

### 1. Получение сертификата (Certbot)
```bash
# Останавливаем nginx временно
docker-compose stop nginx

# Получаем сертификат
sudo docker run --rm \
    -v /opt/qwen_test/nginx/ssl:/etc/letsencrypt \
    -v /opt/qwen_test/nginx/certbot-webroot:/var/www/certbot \
    certbot/certbot certonly \
    --webroot \
    -w /var/www/certbot \
    --email ваш@email.com \
    -d ваш-домен.com \
    -d www.ваш-домен.com
```

### 2. Настройка nginx
```bash
# Редактируем nginx/conf.d/default.conf
nano nginx/conf.d/default.conf

# Изменяем server_name на ваш домен
server_name ваш-домен.com www.ваш-домен.com;
```

### 3. Запуск с SSL
```bash
docker-compose up -d
```

### 4. Авто-обновление сертификатов
```bash
# Certbot уже настроен в docker-compose.yml
# Обновление происходит автоматически каждые 12 часов
```

---

## 📜 Деплой скриптом

### 1. Настройка переменных
```bash
export DEPLOY_USER=root
export DEPLOY_HOST=ваш.server.com
```

### 2. Деплой
```bash
# Локальный запуск
./deploy.sh start

# Деплой на сервер
./deploy.sh deploy

# Просмотр логов
./deploy.sh logs

# Создание бэкапа
./deploy.sh backup
```

### 3. Команды скрипта
```bash
./deploy.sh build     # Сборка Docker образа
./deploy.sh deploy    # Деплой на сервер
./deploy.sh start     # Запуск локально
./deploy.sh stop      # Остановка
./deploy.sh restart   # Перезапуск
./deploy.sh logs      # Логи
./deploy.sh backup    # Бэкап
./deploy.sh help      # Помощь
```

---

## 📊 Мониторинг

### 1. Статус контейнеров
```bash
docker-compose ps
```

### 2. Логи
```bash
# Все сервисы
docker-compose logs -f

# Только приложение
docker-compose logs -f app

# Только nginx
docker-compose logs -f nginx
```

### 3. Использование ресурсов
```bash
docker stats
```

### 4. Health check
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/stats
```

---

## 💾 Бэкапы

### 1. Автоматические бэкапы
Бэкапы создаются каждые 5 минут в `/app/backups/`

### 2. Ручной бэкап
```bash
docker-compose exec app tar -czf /app/backups/backup_$(date +%Y%m%d_%H%M%S).tar.gz /app/data
```

### 3. Копирование бэкапа
```bash
docker cp goquiz-app:/app/backups/backup_20260310_120000.tar.gz ./
```

### 4. Восстановление
```bash
# Копируем бэкап в контейнер
docker cp backup.tar.gz goquiz-app:/app/

# Распаковываем
docker-compose exec app tar -xzf /app/backup.tar.gz -C /
```

---

## 🔧 Troubleshooting

### Приложение не запускается
```bash
# Проверка логов
docker-compose logs app

# Пересборка
docker-compose build --no-cache
docker-compose up -d
```

### Ошибки базы данных
```bash
# Проверка файла БД
docker-compose exec app ls -la /app/data/

# Восстановление из бэкапа
docker-compose exec app tar -xzf /app/backups/latest.tar.gz -C /
```

### Проблемы с SSL
```bash
# Проверка сертификатов
ls -la nginx/ssl/live/

# Принудительное обновление
docker-compose run --rm certbot renew
```

### Nginx не проксирует
```bash
# Проверка конфига
docker-compose exec nginx nginx -t

# Перезагрузка nginx
docker-compose restart nginx
```

### Высокое использование памяти
```bash
# Очистка старых образов
docker image prune -a

# Очистка логов
docker-compose logs --tail=100
```

---

## 📈 Production чеклист

- [ ] Изменён `JWT_SECRET` в `.env`
- [ ] Настроен домен и DNS
- [ ] Получен SSL сертификат
- [ ] Настроен фаервол (UFW)
- [ ] Включён health check
- [ ] Настроены бэкапы
- [ ] Настроен мониторинг логов
- [ ] Протестирован деплой
- [ ] Создан не-root пользователь
- [ ] Включён auto-update (watchtower)

---

## 🔗 Полезные ссылки

- **Docker документация:** https://docs.docker.com/
- **Docker Compose:** https://docs.docker.com/compose/
- **Let's Encrypt:** https://letsencrypt.org/
- **Nginx:** https://nginx.org/

---

**Готово к production деплою!** 🚀
