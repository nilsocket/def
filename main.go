package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nilsocket/def/pkg/db"
	"github.com/nilsocket/def/pkg/vocab"
	"github.com/urfave/cli/v2"
)

var longFlag, synFlag, antFlag, playFlag, rmFlag, cleanDBFlag bool
var dbHomeFlag string

var homeDir, _ = os.UserHomeDir()

var def = &cli.App{
	Name:      "def",
	Usage:     "find definition",
	UsageText: "def [options] [word ...]",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "long", Aliases: []string{"l"}, Usage: "print definition in long format", Destination: &longFlag},
		&cli.BoolFlag{Name: "synonyms", Aliases: []string{"s"}, Usage: "print definitions followed by list of synonms", Destination: &synFlag},
		&cli.BoolFlag{Name: "antonyms", Aliases: []string{"a"}, Usage: "print definitions followed by list of antonyms", Destination: &antFlag},
		&cli.BoolFlag{Name: "playAudio", Aliases: []string{"p"}, Usage: "play audio, if avialable", Destination: &playFlag},
		&cli.BoolFlag{Name: "rm", Aliases: []string{"r"}, Usage: "remove word from database", Destination: &rmFlag},
		&cli.StringFlag{Name: "dbPath", Aliases: []string{"path"}, Value: filepath.Join(homeDir, ".def"), Usage: "path to local database", Destination: &dbHomeFlag},
		&cli.BoolFlag{Name: "cleanDB", Usage: "clean/remove local database", Destination: &cleanDBFlag},
	},
	Before: func(c *cli.Context) error {
		if cleanDBFlag {
			os.RemoveAll(dbHomeFlag)
		}

		db.Open(dbHomeFlag)
		return nil
	},
	Action: defAction,
	After: func(c *cli.Context) error {
		db.Close()
		return nil
	},
	UseShortOptionHandling: true,
}

func main() {
	def.Run(os.Args)
}

func defAction(c *cli.Context) error {
	var words []string

	words = c.Args().Slice()

	// if only `def` is typed, then iterate
	// existing words
	if len(words) == 0 {
		db.Iterate(func(key string) {
			fmt.Println(key)
		})
	} else {
		for _, word := range words {

			vw, ldb := get(word)

			if vw != nil && !ldb {
				if err := db.Put(vw.Word, vw); err != nil {
					log.Println("Put", err)
				}
			}

			if rmFlag {
				if err := db.Del(vw.Word); err != nil {
					log.Println("Del", err)
				}
				// don't print anything
				// while deleting a word
				vw = nil
			}

			if vw != nil {
				printWord(vw)
			}
		}
	}

	return nil
}

func printWord(w *vocab.Word) {
	if longFlag {
		fmt.Print(w.Sprintl())
	}
	if synFlag {
		fmt.Print(w.SprintSyn())
	}
	if antFlag {
		fmt.Print(w.SprintAnt())
	}
	if playFlag {
		w.PlayAudio()
	}

	if !longFlag && !synFlag && !antFlag {
		fmt.Print(w.Sprints())
	}

}

// get word from either database or from internet
//
// `vw.Word`, `Word` field in vw,
// may change if necessary.
// i.e.,
// if user searched for `bharat`
// end result is going to be `Bharat`
// store in db as `Bharat`
// Since, a capatilized word can have different meaning
// Ex: Divine, divine
func get(word string) (*vocab.Word, bool) {

	// If found in DB {
	//     return
	// } else {
	//     If found in Internet {
	//         return
	//    } else {
	//         If suggestion is a capitalWord {
	//             If found in DB {
	//                 return
	//             } else {
	//                 return from Internet
	//             }
	//         } else {
	//             Print Suggestions
	//         }
	//     }
	// }

	vw, sugs, ldb := fetchFromDBorInternet(word)

	if len(sugs) == 0 && vw != nil {
		return vw, ldb
	}

	if capitalizedWord(word, sugs) {
		// if we got suggestion as capitalized word,
		// then it is capital only

		vw, _, ldb := fetchFromDBorInternet(sugs[0])

		if !vw.CapitalOnly {
			vw.CapitalOnly = true
			ldb = false
		}

		return vw, ldb
	}

	fmt.Print(vocab.SprintSuggestions(sugs))

	return nil, false
}

// If found in DB {
//     return
// } else {
//     return from Internet
// }
func fetchFromDBorInternet(word string) (*vocab.Word, []string, bool) {

	vw, err := db.Get(word)                     // ex: `bharat`
	if err != nil && err == db.ErrKeyNotFound { // not found
		capitalWord := strings.Title(strings.ToLower(word))

		vw, err := db.Get(capitalWord) // search for `Bharat`s

		if err != nil && err == db.ErrKeyNotFound {
			vw, sugs := vocab.Get(word) // return from Internet
			return vw, sugs, false
		} else if vw.CapitalOnly { // if it is Bharat only, return
			return vw, nil, true
		} else {
			vw, sugs := vocab.Get(word) // return from Internet
			return vw, sugs, false
		}

	} else if err != nil {
		log.Fatalln(err)
		// we will exit here
	}

	return vw, nil, true
}

// https://en.wikipedia.org/wiki/Capitonym
// Ex: Divine, divine
// Incase of `bharat`, `Bharat` is returned as suggestion
func capitalizedWord(word string, sugs []string) bool {
	if len(sugs) == 1 {
		return strings.ToLower(sugs[0]) == word
	}
	return false
}
