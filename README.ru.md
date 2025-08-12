# Xray Torrent Blocker

[![en](https://img.shields.io/badge/lang-en-red)](https://github.com/kutovoys/xray-torrent-blocker/blob/main/README.md)
[![ru](https://img.shields.io/badge/lang-ru-blue)](https://github.com/kutovoys/xray-torrent-blocker/blob/main/README.ru.md)

Xray Torrent Blocker — это приложение для блокировки использования торрентов пользователями панелей на базе [Xray](https://github.com/XTLS/Xray-core). Приложение анализирует логи, обнаруживает использование торрентов и временно блокирует пользователя, отправляя вебхуки на настроенный URL вебхука.

## Возможности

- Мониторинг логов узлов и панели на предмет использования торрентов
- Блокировка IP-адресов на системном уровне с максимальной скоростью блокировки (Абузы не придут!)
- Прерывание соединений через conntrack - мгновенный разрыв существующих торрент-соединений
- Отправка вебхуков на настроенный URL вебхука
- Настройка через конфигурационный файл
- Поддержка различных файрволов для блокировки (iptables, nftables)
- Настраиваемая продолжительность блокировки
- Поддержка временной блокировки с автоматической разблокировкой
- Установка с помощью пакетных менеджеров apt или yum
- Сохранение состояния блокировки между перезапусками приложения
- Автоматическое восстановление блокировки после перезагрузки системы
- Автоматическая очистка истекших блокировок

## Требования

- Файрвол (iptables или nftables)
- Файл логов Xray с включенным логированием

## Установка

### Скрипт быстрой установки

Самый простой способ установить Xray Torrent Blocker — использовать скрипт установки:

```bash
bash <(curl -fsSL git.new/install)
```

Этот скрипт автоматически:

- Определит архитектуру вашей системы
- Скачает последний релиз
- Установит бинарный файл в `/opt/tblocker/`
- Создаст файл конфигурации по умолчанию
- Настроит systemd сервис
- Запустит сервис

Во время установки вам будет предложено ввести путь к файлу логов и выбрать предпочитаемый файрвол (iptables или nftables). Другие параметры конфигурации можно настроить вручную, отредактировав `/opt/tblocker/config.yaml` при необходимости.

### Из репозитория пакетов

После установки из репозитория будет создана конфигурация по умолчанию в `/opt/tblocker/config.yaml`.

Для базовой работы вам нужно изменить только `LogFile`, указав путь к логам xray.

Также будет создан systemd сервис `tblocker.service` для автоматического запуска при загрузке системы. Автозапуск будет включен. Вам нужно только запустить сервис после редактирования конфигурации:

```bash
systemctl start tblocker
```

#### Системы на основе Debian/Ubuntu

```bash
apt update && apt install -y curl gnupg
curl https://repo.remna.dev/xray-tools/public.gpg | gpg --yes --dearmor -o /usr/share/keyrings/openrepo-xray-tools.gpg
echo "deb [arch=any signed-by=/usr/share/keyrings/openrepo-xray-tools.gpg] https://repo.remna.dev/xray-tools/ stable main" > /etc/apt/sources.list.d/openrepo-xray-tools.list
apt update
apt install tblocker
```

#### Системы на основе RPM

```bash
echo """
[xray-tools-rpm]
name=xray-tools-rpm
baseurl=https://repo.remna.dev/xray-tools-rpm
enabled=1
repo_gpgcheck=1
gpgkey=https://repo.remna.dev/xray-tools-rpm/public.gpg
""" > /etc/yum.repos.d/xray-tools-rpm.repo
yum update
yum install tblocker
```

### Из бинарного файла релизов

1. Установите необходимые зависимости:
   ```bash
   # Для Debian/Ubuntu
   sudo apt install conntrack
   # Для CentOS/RHEL
   sudo yum install conntrack-tools
   ```
2. Скачайте последний релиз для вашей архитектуры с [GitHub Releases](https://github.com/kutovoys/xray-torrent-blocker/releases)
3. Извлеките бинарный файл и сделайте его исполняемым:
   ```bash
   chmod +x tblocker
   ```
4. Переместите в системную директорию:
   ```bash
   sudo mv tblocker /opt/tblocker/
   ```
5. Создайте файл конфигурации `/opt/tblocker/config.yaml` с вашими настройками
6. Скопируйте [файл systemd сервиса](tblocker.service) в `/etc/systemd/system/tblocker.service` и запустите сервис
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable tblocker
   sudo systemctl start tblocker
   ```

## Конфигурация

### Конфигурация по умолчанию

После установки приложение использует конфигурацию по умолчанию. Вы можете настроить её, отредактировав `/opt/tblocker/config.yaml`:

```yaml
# Файл логов для мониторинга
LogFile: "/var/log/remnanode/access.log"

# Продолжительность блокировки в минутах
BlockDuration: 10

# Тег, используемый для идентификации торрент-трафика в логах
TorrentTag: "TORRENT"

# Файрвол для блокировки (iptables, nft)
BlockMode: "iptables"
```

### Расширенная конфигурация

Для продвинутого использования вы можете настроить дополнительные функции:

```yaml
# IP-адреса для обхода блокировки
BypassIPS:
  - "127.0.0.1"
  - "::1"

# Директория для хранения данных о блокировке
StorageDir: "/opt/tblocker"

# Регулярное выражение для обработки имени пользователя в вебхуках
UsernameRegex: "^(.+)$"

# Конфигурация вебхука
SendWebhook: false
WebhookURL: "https://your-webhook-url.com/endpoint"
WebhookTemplate: '{"username":"%s","ip":"%s","server":"%s","action":"%s","duration":%d,"timestamp":"%s"}'
WebhookHeaders:
  Authorization: "Bearer your-token"
  Content-Type: "application/json"
```

## Конфигурация панелей

### Для всех панелей

1. Настройте тегирование bittorrent трафика. Раздел `routing`. Добавьте правило:

   ```json
   {
     "protocol": ["bittorrent"],
     "outboundTag": "TORRENT",
     "type": "field"
   }
   ```

   Здесь `TORRENT` — это тег, который приложение будет использовать для фильтрации логов.

2. Настройте блокировку bittorrent трафика. Раздел `outbounds`. Отправьте весь трафик в blackhole:

   ```json
   {
     "protocol": "blackhole",
     "tag": "TORRENT"
   }
   ```

### Remnawave

1. Создайте директорию для логов:

   ```bash
   mkdir -p /var/log/remnanode
   ```

2. Добавьте volume в `docker-compose.yml` remnanode:

   ```yaml
   volumes:
     - "/var/log/remnanode:/var/log/remnanode"
   ```

3. Настройте логирование в конфигурации xray:

   ```json
   "log": {
       "error": "/var/log/remnanode/error.log",
       "access": "/var/log/remnanode/access.log",
       "loglevel": "error"
   }
   ```

4. Перезапустите remnanode.

### Marzban

1. Создайте директорию для логов:

   ```bash
   mkdir -p /var/lib/marzban-node
   ```

2. Добавьте volume в `docker-compose.yml` marzban-node:

   ```yaml
   volumes:
     - /var/lib/marzban-node:/var/lib/marzban-node
   ```

3. Настройте логирование в конфигурации xray:

   ```json
   "log": {
       "error": "/var/lib/marzban-node/error.log",
       "access": "/var/lib/marzban-node/access.log",
       "loglevel": "error"
   }
   ```

4. Установите значение UsernameRegex в config.yaml:

   ```yaml
   UsernameRegex: "^\\d+\\.(.+)$"
   ```

5. Перезапустите marzban-node.

### Другие панели

Для других панелей на основе Xray убедитесь, что:

1. Файлы логов доступны на хост-системе
2. Формат логов включает необходимую информацию (IP, идентификацию пользователя)
3. Bittorrent трафик правильно помечен в правилах маршрутизации

## Советы

### Работа за TCP прокси

⚠️ **Важно**: Если вы размещаете Nginx/HAProxy/другой TCP прокси перед Xray, убедитесь, что реальный IP клиента передается в Xray через протокол PROXY; иначе вы можете заблокировать 127.0.0.1 или IP вашего сервера вместо реального нарушителя.

**Пример конфигурации Xray:**

```json
{
  "inbounds": [
    {
      "port": 444,
      "protocol": "vless",
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "sockopt": {
          "acceptProxyProtocol": true // принимать PROXY v1/v2 от прокси
        }
      }
    }
  ]
}
```

**Пример конфигурации Nginx:**

```nginx
stream {
    server {
        listen 443;
        proxy_pass 127.0.0.1:444;  # ваш inbound Xray
        proxy_protocol on;         # отправлять PROXY protocol в backend
    }
}
```

**Пример конфигурации HAProxy:**

```
backend xray_backend
    mode tcp
    server xray1 127.0.0.1:444 send-proxy-v2
```

Это гарантирует, что Xray получает реальный IP-адрес клиента в своих логах доступа, позволяя tblocker блокировать правильные IP-адреса.

### Чтение логов

Для чтения логов `tblocker` вы можете использовать следующую команду:

```bash
journalctl -u tblocker -f --no-pager
```

### Конфигурация logrotate

Чтобы предотвратить потребление слишком большого места на диске файлами логов, настройте logrotate:

```bash
sudo bash -c 'cat > /etc/logrotate.d/remnanode <<EOF
/var/log/remnanode/*.log {
    size 50M
    rotate 5
    compress
    missingok
    notifempty
    copytruncate
}
EOF'
```

### Работа с вебхуками

Вебхуки позволяют интегрировать tblocker с внешними системами:

- **Панель**: Включение/отключение пользователя в панели для блокировки на всех узлах
- **Telegram**: Отправка уведомлений в группы Telegram. Для уведомлений администраторов и пользователей.
- **Пользовательские API**: Подключение к вашим собственным системам мониторинга

Для получения вебхуков вы можете использовать [n8n](https://n8n.io/) или любой другой сервис вебхуков.

## Участие в разработке

Мы приветствуем вклад сообщества! Если у вас есть идеи по улучшению или вы нашли ошибку, пожалуйста:

1. Создайте issue на GitHub
2. Сделайте форк репозитория
3. Создайте ветку для функции
4. Внесите изменения
5. Отправьте pull request

Для крупных изменений сначала откройте issue для обсуждения того, что вы хотели бы изменить.

## Рекомендация VPN

Для безопасного и надежного доступа в интернет мы рекомендуем [BlancVPN](https://getblancvpn.com/?ref=tblocker). Используйте промокод `TRYBLANCVPN` для получения 15% скидки на подписку.
