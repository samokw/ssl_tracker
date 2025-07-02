# Go CLI Application to Track SSL Certificate Expiries

## A fully functional terminal application for tracking SSL certificate expiry dates

Features:
- Concurrently check all domains 
- Add and delete domains
- Run individual checks on domains




## Getting Started



Get the binary via curl

```bash
curl -fsSL https://cli.samokw.name/install | bash
```

To build the binary for your system:

```bash
git clone https://github.com/samokw/ssl_tracker.git
```
Follow Go installation instructions and make sure to install at least 1.24.3: https://go.dev/dl/

```bash
cd ssl_tracker/server
```

```bash
go build -o sslcerttop ./cmd/main.go
```

```bash
./sslcerttop
```

How to alias the command so it's available everywhere:

```bash 
echo 'alias sslcerttop="$HOME/path/to/ssl_tracker/server/sslcerttop"' >> ~/.bashrc
```

```bash
source ~/.bashrc
```

```bash
sslcerttop
```

## View Documentation
Tool to view documentation in browser:
```bash 
go install golang.org/x/tools/cmd/godoc@latest
```

Start the documentation server:

```bash 
godoc -http=:8080
```

Then visit http://localhost:8080/pkg/github.com/samokw/ssl_tracker/internal/ssl/

## Next Steps
- Set up binary distribution via Cloudflare
- Automatically check for SSL expiry every 24 hours
- Create a web application port
