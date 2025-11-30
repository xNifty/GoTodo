# GoTodo: A simple web based TODO List app

## Requirements

- PostgreSQL

## Setup and Running

### .env

You will need to rename the `.env.example` file to `.env` and fill in the required information.  
This is needed to connect to the database as well as launch the server. BASE_PATH is used throughout
to handle pathing to ensure things do not break when hosting on something like a subdomain.

```.env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=dbname
SESSION_KEY=32_characters_or_more_string
PORT=8080
BASE_PATH=http://localhost:8080
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

Server available on specified host and port (e.g. localhost:8080).  
List will update automatically when adding a new item, deleting, or toggling complete/incomplete.  
Web makes use of Bootstrap and HTMX for styling and updating elements.

![Screenshot 2024-11-29 152706](https://github.com/user-attachments/assets/da2dce9a-87a1-4982-b07a-9aa46ccbe8bc)

![Screenshot 2024-11-29 152713](https://github.com/user-attachments/assets/43339f33-7fba-4eac-8858-ad02178a9755)

### TODO

- See issue tracker, everything is tracked there at this point

---

Licensed under MIT; see LICENSE file for more information.
