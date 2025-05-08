# Это тестовое задание для Cloud.ru

## Сборка
Для поднятия сервера в докере надо прописать в папке deploy
```
docker-compose build
docker-compose up
```
Сервер поднимется на порту, указанном в configs/config 

## Тесты
Для запусков интеграционных тестов надо запустить следующие команды в корневой папке проекта

```
 go test -race -bench=. -v ./tests/integration    
```

Для PowerShell
```
 go test -race -bench . -v ./tests/integration    
```

