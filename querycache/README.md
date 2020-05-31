# querycache - use cache, save cash

querycache lets you save a query and access them over an HTTP API, caching the results based on a per-query "lifetime" parameter.
It currently supports queries against: Postgres, Snowflake, and MySQL. And addition for more go `sql` compatible drives is easy!

## Adapters

An Adapter is the description of the connection to a specific datasource. It has 3 parameters:

- `name` - a friendly name
- `type` - the driver name (e.g. `postgres`, `mysql`, `snowflake`)
- `options` - the connection string and options.

The `type` and `options` are passed directly to `sql.Open` as the first and second parameter.

The following endpoints are exposed:
- `GET /adapters` - List endpoint, accepts `per` and `page` query parameters
- `POST /adapters` - Create endpoint, accepts json object with `name`, `type`, and `options` keys. (all required)
- `GET /adapters/{id}` - Read endpoint, returns the JSON representation of the adapter
- `PATCH /adapters/{id}` - Update endpoint, accepts json object with `name`, `type`, and `options` keys. (all optional)
- `DELETE /adapters/{id}` - Delete endpoint, deletes the adapter


## Queries

A Query represents a given query to execute.

The following endpoints are exposed:
- `GET /queries` - List endpoint, accepts `per` and `page` query parameters
- `POST /queries` - Create endpoint, accepts json object with `query`, `lifetime`, and `adapterId` keys. (all required)
- `GET /queries/{id}` - Read endpoint, returns the JSON representation of the query
- `PATCH /queries/{id}` - Update endpoint, accepts json object with `query`, `lifetime`, `lastRefresh`, and `adapterId` keys. (all optional)
- `DELETE /queries/{id}` - Delete endpoint, deletes the query


## Examples

The following examples are using `HTTPie`.


Create an adapter:
```
http POST http://localhost:8080/querycache/adapters type=postgres name="PG Local" options="sslmode=disable dbname=querycache"
HTTP/1.1 200 OK
Content-Length: 222
Content-Type: application/json; charset=UTF-8
Date: Sun, 31 May 2020 16:45:50 GMT

{
    "createdAt": "2020-05-31T17:45:50.595626+01:00",
    "id": "b9d444a5-720b-44f9-8ebe-6f07bb02c3e0",
    "name": "PG Local",
    "options": "sslmode=disable dbname=querycache",
    "type": "postgres",
    "updatedAt": "2020-05-31T17:45:50.595627+01:00"
```

Create a query:
```
http POST http://localhost:8080/querycache/queries adapterId=b9d444a5-720b-44f9-8ebe-6f07bb02c3e0 query="SELECT * FROM users" lifetime="8h"         (555ms)
HTTP/1.1 200 OK
Content-Length: 290
Content-Type: application/json; charset=UTF-8
Date: Sun, 31 May 2020 16:46:42 GMT

{
    "adapterId": "b9d444a5-720b-44f9-8ebe-6f07bb02c3e0",
    "createdAt": "2020-05-31T17:46:42.176734+01:00",
    "id": "dc927e8b-8a42-4208-8139-2c1b1af1c243",
    "lastRefresh": "2020-05-31T17:46:42.176734+01:00",
    "lifetime": "8h0m0s",
    "query": "SELECT * FROM users",
    "updatedAt": "2020-05-31T17:46:42.176734+01:00"
}
```

Get query results:
```
http GET http://localhost:8080/querycache/queries/dc927e8b-8a42-4208-8139-2c1b1af1c243/result                                                       (378ms)
HTTP/1.1 200 OK
Content-Length: 70
Content-Type: text/csv
Date: Sun, 31 May 2020 16:49:29 GMT

id,username,name,email
1,christian,Christian Gregg,christian@bissy.io
```
