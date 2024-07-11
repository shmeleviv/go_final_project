# Планировщик задач

Приложение позовляет создавать задачи, привязывать задачи к соответствующим датам, с возможностью автоматического повторения задачи.


## Структура проекта
### ./entities
#### task.go
Содержит основную структур данных, описывающую задачу Task
Описание данных структуры:
- Id - идентификатор задачи в БД
- Date - дата выполнения задачи
- Title - заголовок задачи
- Comment - опциональные комментарии к задаче
- Repeat - правило повтрения задачи

### ./storage
#### storage.go
Содержит структуры и методы для проведения операций c БД
Методы и функции:
- OpenSQLiteDB() (*sqlx.DB, error) - создание БД SQlite  и подключение к ней
- GetAll(time string, limit string) - получить все записи БД  c лимитом limit, date которых >= time
- GetByID(id string) (entities.SchedulerTask, error) - получить запись БД, идентификатор которой равен id
- DeleteByID(id string) error - удалить запись БД, идентификатор которой равен id
- Update(task entities.SchedulerTask) error - обновить запись в БД
- Insert(task entities.SchedulerTask) (string, error) - создать запись в БД



### ./service
#### service.go
Содержит структуры и методы для проведения операция с задачами Task
Методы и функции:
- GetAllTasks() ([]entities.SchedulerTask, error) - получить все ближайшие задачи
- SetTask(task entities.SchedulerTask) (string, error) - создать задачу
- GetTask(id string) (entities.SchedulerTask, error) - получить информацию о существующей задаче
- EditTask(task entities.SchedulerTask) error - редактировать существующую задачу
- DoneTask(id string) error - отметить существующую задачу как выполненную
- DeleteTask(id string) error - удалить задачу

### ./controller
#### handler.go
Содержит обработчики http-запросов 
Методы и функции:
- GetTaskHandler(w http.ResponseWriter, r *http.Request) - обработка http GET /api/task?id=x, использует метод service.GetTask(id string)
- GetTasksHandler(w http.ResponseWriter, r *http.Request) - обработка http GET /api/task, использует метод service.GetAllTasks()
- SetTaskHandler(w http.ResponseWriter, r *http.Request) - обработка http POST /api/task, использует метод service.SetTask(task entities.SchedulerTask)
- EditTaskHandler(w http.ResponseWriter, r *http.Request) - обработка http PUT /api/task, использует метод service.EditTask(task entities.SchedulerTask) 
- DoneTaskHandler(w http.ResponseWriter, r *http.Request) - обработка http POST /api/task/done, использует метод service.DoneTask(id string)
- DeleteTaskHandler(w http.ResponseWriter, r *http.Request) - обработка http DELETE/api/task?id=x, использует метод service.DeleteTask(id string)
- NextDateHandler(w http.ResponseWriter, r *http.Request) - обработка http GET /api/nextdate?now=x,date=x,repeat=x, использует функцию datetime.NextDate()

#### webserver.go
Позволяет запустить web server с соответствующими обработчиками http-запросов


### ./cmd
Содержит файл main.go

### ./datetime
Функции работы с датами

### ./tests
Функции тестирования.
Для запуска:

```bash
go  test -count=1 ./tests
```

### ./web
Содержит файлы с кодом front-end