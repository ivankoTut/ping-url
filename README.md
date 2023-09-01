#### Запуск приложения
развернуть инфраструктуру(бд\логи\редис\тд)
```shell
make up
```
запуск приложения
```shell
make
```
&nbsp;
#### Добавление новой миграции
для postgres
```shell
migrate create -ext sql -dir migrations/postgres create_your_table
```