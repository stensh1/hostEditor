# HostEditor

## Разделы:

- **Краткое описание**
- **Установка и запуск**: О том, как собрать все в работающее приложение.
- **Описание работы модулей**: Детальное описание работы приложения.
  - **Сервер**: Как работает сервер.
  - **Клиент**: Как работает клиент.
- **Контакты**: Связь со мной.

## Краткое описание

Проект предоставляет реализованный на языке Golang GRPC/GRPC-Gateway сервис и клиентское CLI приложение на фреймворке Cobra, для выставления имени хоста в Linux, а также изменения списка DNS серверов.

Сервер работает на задаваемом пользователем порте (стандартно порт 9000), также запускается http proxy сервер (по умолчанию на порте 8080). Сервер способен выполнить команды по отправке информации о текущем hostname, изменинию hostname, отправке информации о списке dns серверов и изменению (добавлению и удалению записей) списка dns серверов. 

Так как для выполнения действий, вносящих изменения в систему, необходимо выполнение ```sudo```, на сервере реализована функция авторизации пользователей. При авторизации сервер запрашивает пароль пользователя и проверяет его на правильность, для этого в переменных окружения хранится md5 сумма пароля пользователя. Если пароль введен верно, сервер генерирует с payload jwt токен и подписывает его на 24 часа, ключ для подписи так же лежит в переменных окружения, после чего отрправляет подписанный токен пользователю и сохраняет запись об авторизации в Redis. Если клиент, при отправке запроса, для которого требуется токен, не находит токен, то автоматически отправляется запрос на авторизацию.

Эндпоинты прокси сервера можно посмотреть в документации Swagger: ``https://stensh1.github.io/hostEditor/``

Все ответы сервера кэшируются в Redis, для соответствия REST, и в первую очередь поиск информации для ответа проводится в базе данных.
  
## Установка и запуск

- ### Самый легкий путь для сборки всего сразу:
  1. Скачайте docker/docker-compose.yaml
  2. Запустите ``docker-compose up --build -d``
  3. Запустится три контейнера: server, client и redis
  4. Чтобы открыть интерактивный режим на клиенте выполните ``docker-compose exec client bash``
  5. В открывшейся оболочке необходимо узнать IP сервера, я делаю это через ``ping server``
  6. Теперь все готово для отправки запросов на сервер. Более подробно о составлении запросов можно почитать в разделе: **Описание работы модулей/Клиент**
 
- ### Сложнее:
  1. Склонируйте репозиторий
  2. Проверьте, что все зависимости подгрузились ``go mod tidy``
  3. Внесите изменения, при необходимости, в конфигурационный файл сервера: /cfg/config.yaml, в подгружаемые переменные окружения: cfg/.env
  4. Из корня проекта /hostEditor соберите образы сервера и клиента:
     ```bash
     docker build -t host-editor-client -f docker/client/Dockerfile .
     docker tag host-editor-client your-dh-name/host-editor-client:latest
     docker push your-dh-name/host-editor-client
     docker build -t host-editor -f docker/server/Dockerfile .
     docker tag host-editor your-dh-name/host-editor:latest
     docker push your-dh-name/host-editor
     ```
  5. Обратите внимание, что /docker/server/Dockerfile создает нового пользователя *user* с заданным паролем и добавляет его в sudoers. В /cfg/.env должна храниться переменная ***USER_PSWD*** с этим захэшированным по MD5 паролем, а также переменная ***ROOT_PSWD*** со значением этого пароля.
  6. *P.S. Я не стал заниматься администрированием пользователей в контейнере и выдачей им различных прав, именно поэтому оставил эти две переменные в /cfg/.env.*
  7. Отредактируйте ***docker/docker-compose.yaml*** с учетом вашего имени на Docker Hub.
  8. Вы также можете заменить пароль для Redis, но не забудьте обновить его в переменных окружения cfg/.env: ***REDIS_PASSWORD***
  9. Выполните шаги, начиная с п. 2 из: ***Самый легкий путь для сборки всего сразу:***

## Описание работы модулей

- ### Сервер
Проект реализован в виде двух отдельных программ: сервера и клиента.

Сервер собирается с помощью библиотеки ***grpc-gateway*** для Golang. Файлы /api/proto/profile.proto и /api/proto/buf.gen.yaml используются для генерации grpc/grpc-gateway кода сервера. Код генерируется с помощью утилиты *buf*. Команды для ее установки есть в Makefile и также ниже:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
	echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bash_profile; \
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
brew install bufbuild/buf/buf
```
Запуск команды для генерации кода: ``cd api/proto/; buf generate``
Команда генерирует файлы: /pkg/api/: profile.pb.go, profile.pb.gw.go, profile_grpc.pb.go и код связанных библиотек /api/proto/google/api и /api/proto/protoc-gen-openaiv2/options, необходимые для генерации документации REST (swagger). Я не стал совмещать всю документацию в один .json файл, поэтому их генерируется три - для нашего profile.proto и для двух зависимостей.

Файл grpcServer.go реализует методы, необходимые для запуска сервера, файл grpcMethods.go имплементирует методы, необходимые для реализации интерфейса, сгенерированного в grpc, файл models.go описывает необходимые структуры.

В процессе запуска создается два экземпляра: http сервер, выступающий в качестве proxy сервера (по умолчанию порт 8080) и grpc сервер (по умолчанию порт 9000)

- ### Клиент
Клиент для подключения к серверу реализован на фреймворке Cobra языка Golang. Частично, код для клиента был сгенерирован командой ``cobra-cli init``. Файл cobraVars.go содержит команды для регистрации в cobra, файл clientMethds содержит методы, которые генерируют запросы к серверу, они передаются в команды cobra, файл root.go сгенерирован, добавлена регистрация новых команд, файл models.go содержит необходимые структуры.

  - ***Реализованные команды***
  ```bash
  ./client s -p=port -u=host [commands] - главная команда для подключения к серверу, требует указания IP адреса и порта сервера
  ./client s -p=port -u=host команды:
    login - запрашивает пароль для авторизации на сервере, запрашивает у сервера jwt токен
    get [command] - получение информации с сервера
    get команды:
      hostname - возвращает hostname сервера
      dns - возвращает список dns серверов, прописанных на сервере
    set [command] [option[ - добавление или изменение информации на сервере
    set команды:
      name [option] - позволяет поменять hostname сервера на [option]
      new_dns [option] - позволяет добавить новую запись [option] в список dns серверов на сервере
      rm_dns [option] - позволяет удалить запись [option] из списока dns серверов на сервере
  ```
  Инструкцию по работе с флагами и командами клиента можно также увидеть, выполнив команду: ``./client s -h`` 


## Контакты
+ tg: t.me/stensh1
+ tel: +7 (921) 982-89-09
+ email: ivan.orshk@gmail.com
