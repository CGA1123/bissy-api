// +heroku goVersion go1.13
// +heroku install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate .
module github.com/cga1123/bissy-api

go 1.13

require (
	github.com/DATA-DOG/go-txdb v0.1.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.0.0-beta.5
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.5.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/honeycombio/beeline-go v0.5.1
	github.com/jcmturner/gokrb5/v8 v8.3.0 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.7.0
	github.com/rs/cors v1.7.0
	github.com/snowflakedb/gosnowflake v1.3.6
	golang.org/x/net v0.0.0-20200528225125-3c3fba18258b // indirect
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
