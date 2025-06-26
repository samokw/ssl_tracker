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
    USER ||--o{ USER_SETTINGS : configures
    DOMAIN ||--o{ NOTIFICATION_LOG : tracks

    USER {
        int id PK
        string email "User's email for notifications"
        string name "Display name for personalized emails"
        datetime created_at "DEFAULT CURRENT_TIMESTAMP"
    }

    USER_SETTINGS {
        int id PK
        int user_id FK "REFERENCES users(id)"
        string setting_key "smtp_host, smtp_port, etc"
        string setting_value "Configuration value"
        datetime created_at "DEFAULT CURRENT_TIMESTAMP"
        datetime updated_at "DEFAULT CURRENT_TIMESTAMP"
    }

    DOMAIN {
        int id PK
        int user_id FK "REFERENCES users(id)"
        string domain_name "NOT NULL"
        datetime created_at "DEFAULT CURRENT_TIMESTAMP"
        datetime last_checked "When SSL was last verified"
        datetime expiry_date "SSL certificate expiry date"
        string last_error "Error message if check failed"
        boolean is_active "DEFAULT 1 - soft delete"
    }

    NOTIFICATION_LOG {
        int id PK
        int domain_id FK "REFERENCES domains(id)"
        int days_before "Notification trigger: 30, 14, 7, 3, 1, 0"
        datetime sent_at "DEFAULT CURRENT_TIMESTAMP"
        string notification_type "email, discord, slack, etc"
    }
```
