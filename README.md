Simple TODO list written in Go. Initial version is in-memory only, and using a CLI for everything.

I plan to expand this as I continue to learn and use Go more. This isn't useful for anything right  
now, but in it's own simple way it works.

## Usage

```go
go run .
```

### Commands:

```
TODO App
1. List Tasks
2. Add Task
3. Complete Task
4. Incomplete Task
5. Delete Task
6. Exit
7. List Current Id
Enter your choice:
```

The commands are straight forward to work with.

On windows, if you go to enter a new task, you may have issues with it immediately jumping to the optional description and therefore be unable to add a task. To work around this, simply enter "2 \<Title Here>" and then when you hit enter it will prompt for description. After this, the task should be added without an issue.
