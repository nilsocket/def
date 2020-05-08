package vocab

import (
	"sync"
)

// Word consists of 4 type of definitions:
// - Short Definition
// - Long Definition
// - Primary Definitions (Links to Full Definitions)
// - Full Definitions
type Word struct {
	Word        string
	Short       string    // .short
	Long        string    // .long
	PrimaryIDs  []string  // tbody a ,redirects to fullDefs, contains id
	FullDefs    []FullDef // .section .definition
	Audios      []Audio
	Examples    []string
	CapitalOnly bool
	wg          *sync.WaitGroup
}

// Audio type alias to raw byte data
type Audio []byte

// FullDef ,full definition of a word
type FullDef struct {
	GroupNum int       // .groupNumber
	Ordinals []Ordinal // div, id
}

// Ordinal ,definitions belonging to
// certain group
type Ordinal struct {
	ID         string     // (id) = sxxxx, ...
	ClassType  string     // (title) = noun, adjective, ...
	Definition string     // .definition
	Examples   []string   // .example
	Instances  []Instance // .instances
}

// Instance in which a word is used
// It usually consists of Type:
// Synonyms, Antonyms, Types, Type Of, ...
type Instance struct {
	Type  string         // {dt}
	Datas []InstanceData // {dd}
}

// InstanceData consists of words for that particular instance
// it's definition if possible
type InstanceData struct {
	Words      []string // .word
	Definition string   // .definition
}

var (
	siteBaseURL  = "https://www.vocabulary.com/dictionary/"
	exBaseURL    = "https://corpus.vocabulary.com/api/1.0/examples.json?query=" // example base url
	audioBaseURL = "https://audio.vocab.com/1.0/us/"
)

// List contains list of examples
type List struct {
	Result *Examples
}

// Examples contains an individual example of word
type Examples struct {
	Sentences []Exsentence
}

// Exsentence contains the actual sentence
type Exsentence struct {
	Sentence string
}
