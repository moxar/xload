// Package xload provides request caching and buffering tools.
/*
Motivation

Request caching and buffering is a common requirement in backend development. As an example, GraphQL performance enhances a lot
when the incomming request uses those two mechanism.

https://github.com/vektah/dataloaden proposes an implementation that generates code to handle both issues.
However, I found that the implementation has limits:

	* it has no native use of request context, though a workaround exists in the FAQ
	* it does not distinguises caching and buffering. Some requests cannot be cached but can be buffered, and vice versa. With dataloaden, you cannot do one without the other.

This package (aims to) provide both logics in an independant, yet combinable way.

SQL example

With two SQL queries A and B

	-- Query A
	SELECT * FROM users WHERE name LIKE "%Clark%";

	-- Query B
	SELECT * FROM users WHERE name LIKE "%Lois%";

	-- Buffered query
	SELECT * FROM users WHERE name LIKE "%Lois%" OR name LIKE "%Clark%";

	-- Note: the buffering strategy matters and may change from one request to another. Some would use UNION others WHERE ... OR, etc.

Benefits of caching and buffering would be:

	* Caching the slice responded by A or B would have little interest. The caching key would be `LIKE "%Clark%"` and the probability it get hits is very low.
	* Caching each row responded by A or B may be interesting, if we decide to use a unique key such as the primary key.
	* Buffering the request is definitely interesting, as it reduces the number of SQL queries executed.

*/
package xload
