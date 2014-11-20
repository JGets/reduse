 #[Redu.se](http://redu.se)
Redu.se is a simple URL shortener, designed to allow the user to reduce the length of a URL that they wish to send to someone, or print out and have people be able to easily type into their web browser.

##Overview
Redu.se can be used to shorten any URL that uses a <code>http://</code> or <code>https://</code> [scheme](http://en.wikipedia.org/wiki/URI_scheme). It requires that the user solve a [CAPTCHA](http://en.wikipedia.org/wiki/CAPTCHA) for every short link they wish to generate, as to dissuade abuse of the service, and prevent spamming & bots from filling the database.

Additionally, Redu.se links can be used to redirect to any resource relative to the shortened URL. For example, consider your website is of the form: <code>hosting.somesite.com/~myusername/</code>. Plug that into [Redu.se](http://redu.se) and it you get a link of the form <code>redu.se/a2c</code>. To direct someone to your "About" page which is located at <code>hosting.somesite.com/~myusername/about/aboutme.html</code>, you can simply use <code>redu.se/a2c/about/aboutme.html</code>.

Similarly, Redu.se treats GET parameters the same way. Consider you want to link to <code>hosting.somesite.com/~myusername/?title=someTitle&date=today</code>. Simply using <code>redu.se/a2c/?title=someTitle&date=today</code> would give you the same result.

##About
Redu.se is written entirely in the [Go](http://golang.org) programming language, and makes use of the following 3rd party packages:
* [Web.go](http://webgo.io) for a simplified web app framework
* [Captcha](http://github.com/dchest/captcha) for generating and verifying [CAPTCHA](http://en.wikipedia.org/wiki/CAPTCHA)s to prevent spamming
<!--- * [Go-MySQL-Driver](github.com/go-sql-driver/mysql) as a MySQL Driver for Go's [database/sql](http://golang.org/pkg/database/sql) package -->
* [pq](github.com/lib/pq) as a PostgreSQL driver for Go's [database/sql](http://golang.org/pkg/database/sql) package
* [GoRelic](http://github.com/yvasiyarov/gorelic) for integration with [NewRelic](http://newrelic.com) for app performance statistics

It uses a <!--- MySQL --> PostgreSQL database back-end to store the short links and their URL destinations.

Currently, no click statistics are kept, but this may be implemented in the future. However, if and when any statistics are gathered, it will be completely anonymous, and available to the public (statistics for specific links will __not__ be gathered, but information about the number of short links used per day will be gathered).
