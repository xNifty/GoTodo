# GoTodo: A simple CLI App with a basic web front end

## Requirements
- PostgreSQL

## Setup and Running

### .env
You will need to rename the ```.env.example``` file to ```.env``` and fill in the required information. This is needed to connect to the database.

### To start the CLI app:
```go
go run .
```

### To start the web app:
```go
go run . server
```

Server available on localhost:8080 and refreshes every 10 seconds.

---

The commands are straight forward to work with.

On windows, if you go to enter a new task, you may have issues with it immediately jumping to the optional description and therefore be unable to add a task. To work around this, simply enter "2 \<Title Here>" and then when you hit enter it will prompt for description. After this, the task should be added without an issue.
