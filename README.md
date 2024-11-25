# GoTodo: A simple web based TODO List app

## Requirements
- PostgreSQL

## Setup and Running

### .env
You will need to rename the ```.env.example``` file to ```.env``` and fill in the required information. This is needed to connect to the database.

### To start:
```go
go run .
```

Server available on localhost:8080.  
List will update automatically when adding a new item, deleting, or toggling complete/incomplete.  
Web makes use of Bootstrap and HTMX for styling and updating elements.

---

Licensed under MIT; see LICENSE file for more information.

