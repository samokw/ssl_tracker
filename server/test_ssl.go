package main

import (
	"fmt"
	"log"

	"github.com/samokw/ssl_tracker/internal/database"
	"github.com/samokw/ssl_tracker/internal/domain"
	"github.com/samokw/ssl_tracker/internal/ssl"
	"github.com/samokw/ssl_tracker/internal/types"
)

func main() {
	dbPath, err := database.GetDefaultDBPath()
	if err != nil {
		log.Fatal(err)
	}
	
	db, err := database.InitSQLite(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	domainRepo := domain.NewRepository(db)
	sslService := ssl.NewCertService()
	domainService := domain.NewService(domainRepo, sslService)
	
	fmt.Println("Testing SSL checking for all domains...")
	err = domainService.CheckAllDomainsSSLSync(types.UserID(1))
	if err != nil {
		log.Printf("Error checking SSL: %v", err)
	}
	
	domains, err := domainService.GetUsersDomains(types.UserID(1))
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("\nResults for %d domains:\n", len(domains))
	for _, d := range domains {
		fmt.Printf("Domain: %s\n", d.DomainName.String())
		if d.ExpiryDate != nil {
			fmt.Printf("  Expires: %s\n", d.ExpiryDate.Time().Format("2006-01-02"))
		} else {
			fmt.Printf("  Expires: Unknown\n")
		}
		if d.LastError != nil {
			fmt.Printf("  Error: %s\n", d.LastError.String())
		}
		if d.LastChecked != nil {
			fmt.Printf("  Last checked: %s\n", d.LastChecked.Time().Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}
}
