package dao

// import ()

type Entity struct {
    Name string
    Path string
    User string
    Host string
    Type string
}

type EntityList struct {
    Type string
    Entities []Entity
}
