# Это тестовое задание для Cloud.ru

## Сборка
Для поднятия сервера в докере надо прописать в папке deploy
```
docker-compose build
docker-compose up
```
Сервер поднимется на порту, указанном в configs/config 

Чтобы добавить особые лимиты для клиента, надо сделать запрос через /clients методом POST в котором лежит информация в следующем виде
```
{
   "client_ip": string,
   "capacity": number,
   "rate_per_sec": number
}
```

Чтобы удалить особые лимиты для клиента, надо сделать запрос через /clients/{clientIp} методом DELETE

## Тесты
Для запусков интеграционных тестов надо запустить следующие команды в корневой папке проекта

```
 go test -race -bench=. -v ./tests/integration    
```

Для PowerShell
```
 go test -race -bench . -v ./tests/integration    
```

