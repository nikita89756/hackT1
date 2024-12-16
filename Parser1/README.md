# Universal Document Parser Service

## Описание
Этот проект представляет собой универсальный сервис для парсинга различных типов документов с REST API интерфейсом. Сервис поддерживает:
- PDF документы
- DOCX файлы
- Веб-страницы (URL)
- Текстовые файлы

Основная логика парсинга реализована на Python с использованием FastAPI для REST API интерфейса.

## Системные требования
- Python 3.9+
- Docker (опционально)

## Установка и настройка

### 1. Подготовка Python окружения
```bash
# Создание виртуального окружения
python -m venv .venv

# Активация окружения
# Для Windows:
.venv\Scripts\activate
# Для Linux/MacOS:
source .venv/bin/activate

# Установка зависимостей
pip install -r requirements.txt
```

### 2. Запуск сервера
```bash
python api.py
```
Сервер будет доступен по адресу: http://localhost:8000

### Docker установка
```bash
# Сборка образа
docker build -t document-parser .

# Запуск контейнера
docker run -p 8000:8000 document-parser
```

## API Endpoints

### 1. Парсинг документа по URL или локальному пути
```http
POST /parse
Content-Type: application/json

{
    "source": "https://example.com/document.pdf",
    "output_path": "optional/path/to/save.txt"
}
```

### 2. Парсинг загруженного PDF файла
```http
POST /parse_pdf
Content-Type: multipart/form-data

file: document.pdf
```

### 3. Парсинг загруженного DOCX файла
```http
POST /parse_docx
Content-Type: multipart/form-data

file: document.docx
```

### 4. Проверка работоспособности сервиса
```http
GET /health
```

## Особенности
- Поддержка различных форматов документов (PDF, DOCX, веб-страницы)
- Асинхронная обработка запросов
- Опциональное сохранение результатов в файл
- Docker поддержка для простого развертывания
- Обработка таблиц в DOCX документах
- Извлечение текста из многостраничных PDF документов

## Обработка ошибок
Сервис предоставляет подробные сообщения об ошибках в случае:
- Неподдерживаемого формата файла
- Недоступности URL
- Ошибок при парсинге документа
- Проблем с сохранением результатов

## Безопасность
- Валидация входных данных
- Безопасная обработка файлов
- Временные файлы автоматически удаляются