# Xload

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/moxar/xload)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/moxar/xload/master/LICENSE)

`xload` is a Golang request caching and buffering toolkit.

## Motivation

Request caching and buffering is a common requirement in backend development. As an example, GraphQL performance enhances a lot
when the incomming request uses those two mechanism.

[dataloader](https://github.com/vektah/dataloaden) proposes an implementation that generates code to handle both issues.
However, the implementation does not distinguises caching and buffering. Some requests cannot be cached but can be buffered, and vice versa.
This package (aims to) provide both features in an independant, yet combinable way.

### SQL demonstration
```SQL

-- Query A
SELECT * FROM users WHERE name LIKE "%Clark%";

-- Query B
SELECT * FROM users WHERE name LIKE "%Lois%";

-- Buffered query
SELECT * FROM users WHERE name LIKE "%Lois%" OR name LIKE "%Clark%";
```

* Caching the slice responded by A or B would have little interest. The caching key would be `LIKE "%Clark%"` and the probability it get hits is very low.
* Caching each row responded by A or B may be interesting, if we decide to use a unique key such as the primary key.
* Buffering the request is definitely interesting, as it reduces the number of SQL queries executed.

_Note: the buffering strategy matters and may change from one request to another. Some would use `UNION`, other `WHERE ... OR`, etc._

Finaly, the initialization of the buffer or cache must be cheap. They are meant to be scoped to a unique job such as a http request.

## Definitions

The following definitions must be evaluated in regard of this package's context.
Note that the request is agnostic of the backend (REST, HTTP, SQL, Redis...).

### Buffering

Buffering consists in aggregating similar requests emitted during a short period of time, to send a unique, bigger request.

### Caching

Caching consists in saving, in memory, the response of a request for a short period of time. This response must be accessible with
a unique identification that is tied to the request.
