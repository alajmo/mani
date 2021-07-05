package print

import (
	"fmt"
)

func PrintTags(tags []string) {
	for _, tag := range tags {
		fmt.Println(tag)
	}
}


