// +heroku goVersion go1.14
// +heroku install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate .
module github.com/cga1123/bissy-api

go 1.14

require (
	github.com/DATA-DOG/go-txdb v0.1.4
	github.com/PagerDuty/go-pagerduty v1.4.1
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bugsnag/bugsnag-go v1.9.0
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.6.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/honeycombio/beeline-go v0.11.1
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.2
	github.com/rs/cors v1.7.0
	github.com/slack-go/slack v0.9.1
	github.com/snowflakedb/gosnowflake v1.5.0
)
