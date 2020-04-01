# Textbroker Author Article Picker

Tired of refreshing the textbroker's author page to find for new articles to write and to miss them ?

This script/bot does:
Automatically select an order when available as an textbroker author.
When selected, the order is blocked for you (textbroker mechanism), and you will have 10 minutes to go on textbroker and validate if you want it or not.
This program will also make a loud sound so that you are notified when an order is selected.

For author textbroker only, not clients !
This is basically a textbroker author orders scrappers, with a select functionality too.

# Features

- You can set the minimum gain in euros that you want for the order.
- Do not select order that you have already selected before.

# Installation

Very simple :

```
git clone git@github.com:kangoo13/textbroker-author-article-picker.git
cd textbroker-author-article-picker
cp .env.dist .env
```

At this point, do not forget to modify the values in the .env file. (login and password to connect on textbroker mostly)
And finally:

```
go build
./textbroker-author-article-picker
```

# Contribution

Any enhancements or even ideas are welcome, through issues !

# Chat

Feel free to speak with me on FreeNode IRC, my nickname is kangoo13
