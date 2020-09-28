# gorm-paginator

## Installation

```bash
go get -u github.com/mmuflih/gorm-paginator
```

## Example Response
```json
"data": [
    {
        "id": 1,
        "name": "Muflih Kholidin",
        "phone": "xxxxxxxxxxxxxx",
    },
],
"paginate": {
    "page": 1,
    "size": 10,
    "total": 3
}
```

## Usage

```go
import (
    paginator "github.com/mmuflih/gorm-paginator"
)

type Student struct {
	ID       int
	Name     string
	Phone  string
}

var items []Student
db = db.Where("name like ?", "%muf%")

paginator.Make(&paginator.Config{
    DB:      db,
    Page:    1,
    Size:   10,
    OrderBy: []string{"name asc"},
}, &items)
```