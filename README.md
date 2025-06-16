# Печенев Платон, БПИ-2310

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/gin-%23008ECF.svg?style=for-the-badge&logo=gin&logoColor=white)
![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)
![Postman](https://img.shields.io/badge/Postman-FF6C37?style=for-the-badge&logo=postman&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![Nginx](https://img.shields.io/badge/nginx-%23009639.svg?style=for-the-badge&logo=nginx&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/Rabbitmq-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)

[Postman коллекция](./collection.postman_collection.json)

## Особенности

Для передачи событий используется Protobuf. Код генерится с помощью `protoc` и `protoc-gen-go` внутри контейнера.

В payments реализованы transactional inbox (`order_created`) и outbox (`order_success` / `order_failed`).

В order используется transactional outbox (`order_created`). События `order_success` и `order_failed` слушаются воркером `tasks`.
