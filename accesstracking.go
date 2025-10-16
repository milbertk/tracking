package tracking

import (
	"fmt"

	"github.com/milbertk/databasesmng"
)

// Insert inserts the FirebaseUser into the usfirebasedata table
func (lt *LoginTracking) Insert() error {
	db, err := databasesmng.CreateConnection()
	if err != nil {
		return fmt.Errorf("❌ DB connection error: %v", err)
	}

	query := `
		INSERT INTO public.logintracking (
	userid, email, datelocalacces, ip, platform,
	macaddress, browser, countrycode, gmttime, lang, action, jsonstring
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`

	_, err = db.Exec(query,
		lt.UserID,
		lt.Email,
		lt.DateLocalAcces,
		lt.IP,
		lt.Platform,
		lt.MacAddress,
		lt.Browser,
		lt.CountryCode,
		lt.GMTTime,
		lt.Lang,
		lt.Action,
		lt.Jsonstring,
	)

	if err != nil {
		return fmt.Errorf("❌ Failed to insert user: %v", err)
	}

	fmt.Println("✅ Tracking inserted succesfully")
	return nil
}
