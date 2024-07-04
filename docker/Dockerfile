# Используем официальный образ golang на базе alpine в качестве базового образа
FROM golang:alpine

# Устанавливаем sudo
RUN apk update && apk add sudo

# Создаем нового пользователя user с паролем admin
RUN adduser -D user && echo "user:admin" | chpasswd

# Добавляем пользователя user в sudoers
RUN echo "user ALL=(ALL) ALL" >> /etc/sudoers

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируем все файлы проекта, исключая .bin, cmd/client, docker, logs, .gitignore, dump.rdb, makefile, readme.md
COPY . .
RUN rm -rf .bin cmd/client docker logs .gitignore dump.rdb Makefile README.md

# Сборка приложения
RUN go build -o server ./cmd/server

# Устанавливаем права для запуска от имени пользователя user
RUN chown -R user:user /app

# Запуск приложения от имени пользователя user
USER user

# Указываем команду запуска
CMD ["./server"]