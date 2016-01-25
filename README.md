# simple-go-blog

A simple blog written in Go with [Bootstrap](https://getbootstrap.com) for personal use and as a teaching experience.
The blog can be seen live at [rbjorklin.com](https://rbjorklin.com). If someone would want to use this that would require dumping some [Bootstrap](https://getbootstrap.com)
assets under a 'static' folder, more detailed instructions might come.

## Short intro

* The webserver will look after about.html in the same folder as the webserver is run from and publish it under about.
* The webserver will look after posts under posts/\<year\>/\<month\>/\<day\>/\<post_name\> and publish a list under the posts nav-item.
* The webserver will always show the most recent post on the start page.
* The webserver will write a config.json on the first run which you probably want to edit.

## TODO

I need to cleanup the templates folder. A generic template should be more than enough.
