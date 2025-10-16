package tracking

type LoginTracking struct {
	UserID         string `db:"userid"`
	Email          string `db:"email"`
	DateLocalAcces string `db:"datelocalacces"`
	IP             string `db:"ip"`
	Platform       string `db:"platform"`
	MacAddress     string `db:"macaddress"`
	Browser        string `db:"browser"`
	CountryCode    string `db:"countrycode"`
	GMTTime        string `db:"gmttime"`
	Lang           string `db:"lang"`
	Action         string `db:"action"`
	Jsonstring     string `db:"jsonstring"`
}
