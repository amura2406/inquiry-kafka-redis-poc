package faker

import (
	"fmt"
	"math/rand"
	"strings"
)

var lorem DataFaker
var wordList = []string{
	"alias", "consequatur", "aut", "perferendis", "sit", "voluptatem",
	"accusantium", "doloremque", "aperiam", "eaque", "ipsa", "quae", "ab",
	"illo", "inventore", "veritatis", "et", "quasi", "architecto",
	"beatae", "vitae", "dicta", "sunt", "explicabo", "aspernatur", "aut",
	"odit", "aut", "fugit", "sed", "quia", "consequuntur", "magni",
	"dolores", "eos", "qui", "ratione", "voluptatem", "sequi", "nesciunt",
	"neque", "dolorem", "ipsum", "quia", "dolor", "sit", "amet",
	"consectetur", "adipisci", "velit", "sed", "quia", "non", "numquam",
	"eius", "modi", "tempora", "incidunt", "ut", "labore", "et", "dolore",
	"magnam", "aliquam", "quaerat", "voluptatem", "ut", "enim", "ad",
	"minima", "veniam", "quis", "nostrum", "exercitationem", "ullam",
	"corporis", "nemo", "enim", "ipsam", "voluptatem", "quia", "voluptas",
	"sit", "suscipit", "laboriosam", "nisi", "ut", "aliquid", "ex", "ea",
	"commodi", "consequatur", "quis", "autem", "vel", "eum", "iure",
	"reprehenderit", "qui", "in", "ea", "voluptate", "velit", "esse",
	"quam", "nihil", "molestiae", "et", "iusto", "odio", "dignissimos",
	"ducimus", "qui", "blanditiis", "praesentium", "laudantium", "totam",
	"rem", "voluptatum", "deleniti", "atque", "corrupti", "quos",
	"dolores", "et", "quas", "molestias", "excepturi", "sint",
	"occaecati", "cupiditate", "non", "provident", "sed", "ut",
	"perspiciatis", "unde", "omnis", "iste", "natus", "error",
	"similique", "sunt", "in", "culpa", "qui", "officia", "deserunt",
	"mollitia", "animi", "id", "est", "laborum", "et", "dolorum", "fuga",
	"et", "harum", "quidem", "rerum", "facilis", "est", "et", "expedita",
	"distinctio", "nam", "libero", "tempore", "cum", "soluta", "nobis",
	"est", "eligendi", "optio", "cumque", "nihil", "impedit", "quo",
	"porro", "quisquam", "est", "qui", "minus", "id", "quod", "maxime",
	"placeat", "facere", "possimus", "omnis", "voluptas", "assumenda",
	"est", "omnis", "dolor", "repellendus", "temporibus", "autem",
	"quibusdam", "et", "aut", "consequatur", "vel", "illum", "qui",
	"dolorem", "eum", "fugiat", "quo", "voluptas", "nulla", "pariatur",
	"at", "vero", "eos", "et", "accusamus", "officiis", "debitis", "aut",
	"rerum", "necessitatibus", "saepe", "eveniet", "ut", "et",
	"voluptates", "repudiandae", "sint", "et", "molestiae", "non",
	"recusandae", "itaque", "earum", "rerum", "hic", "tenetur", "a",
	"sapiente", "delectus", "ut", "aut", "reiciendis", "voluptatibus",
	"maiores", "doloribus", "asperiores", "repellat",
}

// DataFaker generates randomized Words, Sentences and Paragraphs
type DataFaker interface {
	Word() string
	Sentence() string
	Paragraph() string
}

// SetDataFaker sets Custom data in lorem
func SetDataFaker(d DataFaker) {
	lorem = d
}

// GetLorem returns a new DataFaker interface of Lorem struct
func GetLorem() DataFaker {
	mu.Lock()
	defer mu.Unlock()

	if lorem == nil {
		lorem = &Lorem{}
	}
	return lorem
}

// Lorem struct
type Lorem struct {
}

// Word returns a word from the wordList const
func (l Lorem) Word() string {
	return randomElementFromSliceString(wordList)
}

// Sentence returns a sentence using the wordList const
func (l Lorem) Sentence() (sentence string) {
	r, _ := RandomInt(1, 6)
	size := len(r)
	for key, val := range r {
		if key == 0 {
			sentence += strings.Title(wordList[val])
		} else {
			sentence += wordList[val]
		}
		if key != size-1 {
			sentence += " "
		}
	}
	return fmt.Sprintf("%s.", sentence)
}

// Paragraph returns a series of sentences as a paragraph using the wordList const
func (l Lorem) Paragraph() (paragraph string) {
	size := rand.Intn(10) + 1
	for i := 0; i < size; i++ {
		paragraph += l.Sentence()
		if i != size-1 {
			paragraph += " "
		}
	}
	return paragraph
}
