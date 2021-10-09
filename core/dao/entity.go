package dao

// import ()

type Entity struct {
    Name string
    Path string
    Type string
}

type EntityList struct {
    Type string
    Entities []Entity
}
