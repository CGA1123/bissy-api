// +build tools

package tools

import (
	// importing to allow building on migrate cli on heroku
	_ "github.com/golang-migrate/migrate/v4"
	// importing to allow building on migrate cli on heroku
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
