package dao

// import ()

type Entity struct {
    Name string
    Path string
    Type string
}

type EntityList struct {
    Type     string
    Entities []Entity
}

func (e EntityList) GetLongestNameLength() int {
    max := 0
    for _, entity := range e.Entities {
	nameLength := len(entity.Name)
	if (nameLength > max) {
	    max = nameLength
	}
    }

    return max
}
