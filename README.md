# GoTodo: A simple web based TODO List app

## Requirements
- PostgreSQL

## Setup and Running

### .env
You will need to rename the ```.env.example``` file to ```.env``` and fill in the required information. This is needed to connect to the database.

```.env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=dbname
```

### Current Features
- Add a new item to the list
- Delete an item from the list
- Mark an item as complete or incomplete
- List updates automatically when adding, deleting, or toggling complete/incomplete

### To start:
```go
go run .
```

Server available on localhost:8080.  
List will update automatically when adding a new item, deleting, or toggling complete/incomplete.  
Web makes use of Bootstrap and HTMX for styling and updating elements.

![Screenshot 2024-11-29 152706](https://github.com/user-attachments/assets/da2dce9a-87a1-4982-b07a-9aa46ccbe8bc)

![Screenshot 2024-11-29 152713](https://github.com/user-attachments/assets/43339f33-7fba-4eac-8858-ad02178a9755)

### TODO
- Add filtering by date and/or title
- Add searching by title and/or description

---

Licensed under MIT; see LICENSE file for more information.

