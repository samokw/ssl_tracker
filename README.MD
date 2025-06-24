## Getting Started

```bash
git clone https://github.com/samokw/ssl_tracker.git
```
follow go installation instructions and make sure to install at least 1.24.3 : https://go.dev/dl/


## View documentation
Tool to view documentation in browser
```bash 
go install golang.org/x/tools/cmd/godoc@latest
```
Start the documentation server

```bash 
godoc -http=:8080
```

Then visit http://localhost:8080/pkg/github.com/samokw/ssl_tracker/internal/ssl/


```mermaid
erDiagram
    USER ||--o{ DOMAIN : owns
    DOMAIN ||--o{ NOTIFICATION : sends

    USER {
        int id PK
        string email
        string password
        string discord_id
    }

    DOMAIN {
        int id PK
        int user_id FK
        string domain_name
        datetime created_at
        date expiry_date
    }

    NOTIFICATION {
        int id PK
        int domain_id FK
        int days_before
        datetime sent_at
    }

```