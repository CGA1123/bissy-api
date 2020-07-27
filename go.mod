// +heroku goVersion go1.13
// +heroku install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate .
module github.com/cga1123/bissy-api

go 1.13

require (
	github.com/DATA-DOG/go-txdb v0.1.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.0.0-beta.7
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang-migrate/migrate/v4 v4.12.0
	github.com/google/go-cmp v0.5.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/honeycombio/beeline-go v0.5.1
	github.com/jcmturner/gokrb5/v8 v8.3.0 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.7.1
	github.com/rs/cors v1.7.0
	github.com/snowflakedb/gosnowflake v1.3.6
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
