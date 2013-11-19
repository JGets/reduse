#[Redu.se](http://redu.se)
Redu.se is a simple URL shortener, designed to allow the user to reduce the length of a URL that they wish to send to someone, or print out and have people be able to easily type into their web browser.

##About
Redu.se is written entirely in the [Go](http://golang.org) programming language, and makes use of the following 3rd party packages:
* [Web.go](http://webgo.io) for a simplified web app framework
* [Captcha](http://github.com/dchest/captcha) for generating and verifing [CAPTCHA](http://en.wikipedia.org/wiki/CAPTCHA)s to prevent spamming
* [Go-MySQL-Driver](github.com/go-sql-driver/mysql) as a MySQL Driver for Go's [database/sql](http://golang.org/pkg/database/sql) package
* [GoRelic](http://github.com/yvasiyarov/gorelic) for integration with [NewRelic](http://newrelic.com) for app performance statistics

It uses a MySQL database backend to store the short links and their URL destinations.

Currently, no click statistics a kept, but this is planned to be implemented in the future. However, if and when any statistics are gathered, it will be completely anonymous, and available to the public (statistics for specific links will __not__ be gathered, but information link the number of short links used per day will be gathered).

##Overview
Redu.se can be used to shorten any URL that uses a <code>http://</code>, <code>https://</code>, or <code>mailto:</code> [scheme](http://en.wikipedia.org/wiki/URI_scheme). It requires that the user solve a [CAPTCHA](http://en.wikipedia.org/wiki/CAPTCHA) for every short link they wish to generate, as to dissuade abuse of the service, and prevent spamming & bots from filling the database.

Additionaly, Redu.se links can be used to redirect to any resource relative to the shortened URL. For example, consider your website is of the form: <code>http://hosting.somesite.com/~myusername/</code>. Plug that into [Redu.se](http://redu.se) and it you get a link of the form <code>redu.se/a2c</code>. To direct someone to your "About" page which is located at <code>http://hosting.somesite.com/~myusername/about/aboutme.html</code>, you can simply use <code>redu.se/a2c/about/aboutme.html</code>.

Similarily, Redu.se treats GET parameters the same way. Consider you want to link to <code>http://hosting.somesite.com/~myusername/?title=someTitle&date=today</code>. Simply using <code>redu.se/a2c/?title=someTitle&date=today</code> would give you the same result.

More information to come, this is still an early work in progress
