synccomments
------------
Copy podio comments from one app to another.

A bit raw around the edges, but with a bit of luck you might get it to
work for you.

    $ synccomments -h
    Usage of synccomments:
        -clientid="": client id
        -clientsecret="": client secret
        -f=false: force comment inclusion even if preflight checks fail
        -from=0: the app id to move comments from
        -to=0: the app id to move comments to
        -totoken="": app token from the receiving app
    Client settings can be found in Account Settings -> API Key
    .
    App information can be found in the individual App settings
    dropdown menu -> Developer
    .
    You must supply all flags


To install
    go get github.com/stengaard/synccomments
