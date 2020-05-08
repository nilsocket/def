package vocab

import (
	"fmt"
	"strconv"
	"strings"
)

// Indent for printing
var Indent = "   "

// Sprints ,returns shortly expanded string
func (w *Word) Sprints() string {
	return w.printWord("short")
}

// Sprintl ,returns long-expanded string
func (w *Word) Sprintl() string {
	return w.printWord("long")
}

// SprintSyn ,returns a string of synonyms
func (w *Word) SprintSyn() string {
	return w.sprintSynAnt("Synonyms")
}

// SprintAnt ,returns a string of antonyms
func (w *Word) SprintAnt() string {
	return w.sprintSynAnt("Antonyms")
}

// printWord ,prints given word,
// printType = "short" || "long"
func (w *Word) printWord(printType string) string {
	b := &strings.Builder{}

	// short definition
	if w.Short != "" {
		b.WriteString(w.Short)
		b.WriteString("\n\n")
	}

	// long definition
	if w.Long != "" {
		b.WriteString(w.Long)
		b.WriteString("\n\n")
	}

	if printType == "short" {

		// if primary definitions are not available
		// print 3 full definitions, if avialable
		if !pPrimaryDefs(w, b) {
			pDefs(w, 3, b)
		}

		if len(w.Examples) > 3 {
			pExamples(w.Examples[:3], Indent, b)
		} else if len(w.Examples) > 0 {
			pExamples(w.Examples, Indent, b)
		}

	} else if printType == "long" {
		pFullDefs(w, len(w.FullDefs), b)
		pExamples(w.Examples, Indent, b)
	}

	return b.String()
}

func pPrimaryDefs(w *Word, b *strings.Builder) bool {
	// if primaryIDs exist, FullDefs exist
	if w.PrimaryIDs != nil {

		b.WriteString("Definitions:\n")

		prevGroupNum := 0

		i := 1
		for _, pID := range w.PrimaryIDs {
			for _, fdef := range w.FullDefs {
				for _, ord := range fdef.Ordinals {
					if ord.ID == pID { // matched
						if prevGroupNum != fdef.GroupNum {
							b.WriteString("\n" + strconv.Itoa(fdef.GroupNum) + "\n")
							prevGroupNum = fdef.GroupNum
							i = 1
						}
						pDefinition(ord, i, b)
						i++
					}
				}
			}
		}
		return true
	}

	return false
}

// pDefinition ,prints definition along with class type
func pDefinition(ord Ordinal, id int, b *strings.Builder) {
	b.WriteString(
		Indent +
			strconv.Itoa(id) +
			". [" +
			ord.ClassType +
			"] " +
			ord.Definition +
			"\n",
	)
}

// TODO: This can be improved
// If we have 2 ordinals, and each contain 4 definitions
// It would be better if we printed 2 from the first one,
// and third from the second one
func pDefs(w *Word, count int, b *strings.Builder) {
	if w.FullDefs != nil {
		b.WriteString("\nDefinitions:\n")

		overAllCount := 0

		for i := 0; i < count && i < len(w.FullDefs); i++ {
			fullDef := w.FullDefs[i]

			b.WriteString("\n" + strconv.Itoa(fullDef.GroupNum) + "\n")

			for oid, ord := range fullDef.Ordinals {

				overAllCount++
				if overAllCount > count {
					break
				}

				pDefinition(ord, oid+1, b) // definition
			}

			if overAllCount > count {
				break
			}
		}
	}
}

func pFullDefs(w *Word, count int, b *strings.Builder) {
	if w.FullDefs != nil {
		b.WriteString("\nDefinitions:\n")

		for i := 0; i < count && i < len(w.FullDefs); i++ {
			fullDef := w.FullDefs[i]

			b.WriteString("\n" + strconv.Itoa(fullDef.GroupNum) + "\n")

			for oid, ord := range fullDef.Ordinals {
				pDefinition(ord, oid+1, b)            // definition
				pFullExamples(ord.Examples, b)        // full example
				printInstances(ord.Instances, b, nil) // instances
			}
		}
	}
}

// Section deals with printing Instances

// pInsOpts ,print instance options
// instance options for printing
type pInsOpts struct {
	Type                   string // matched type will be printed
	Words                  bool   // print words
	Data                   bool   // print definition
	TypeIndent, WordIndent int
}

func printInstances(instances []Instance, b *strings.Builder, insOpts *pInsOpts) {

	// if opts is nil, create with default opts
	if insOpts == nil {

		insOpts = &pInsOpts{Words: true, Data: true, TypeIndent: 2, WordIndent: 3}
	}

	for _, ins := range instances {
		printInstance(ins, b, insOpts)
	}
}

func printInstance(ins Instance, b *strings.Builder, insOpts *pInsOpts) {

	typeIndent := strings.Repeat(Indent, insOpts.TypeIndent)
	wordIndent := strings.Repeat(Indent, insOpts.WordIndent)

	// if we have data print it's type
	if len(ins.Datas) != 0 {
		if insOpts.Type == ins.Type || insOpts.Type == "" {

			if ins.Type != "" {
				b.WriteString(typeIndent + ins.Type + ":\n")
			}

			for i, insData := range ins.Datas {
				if i == 0 {
					printInsData(insData, b, insOpts, wordIndent, wordIndent)
				} else {
					if !insOpts.Data {
						printInsData(insData, b, insOpts, ", ", wordIndent)
					} else {
						printInsData(insData, b, insOpts, "\n"+wordIndent, wordIndent)
					}
				}
			}

			b.WriteString("\n\n")
		}
	}
}

func printInsData(insData InstanceData, b *strings.Builder, insOpts *pInsOpts, wordIndent, defIndent string) {
	if insOpts.Words {
		for i, word := range insData.Words {
			if i == 0 {
				b.WriteString(wordIndent + word)
			} else {
				b.WriteString(", " + word)
			}
		}
	}

	if insOpts.Data {
		// definition or explanation
		if insData.Definition != "" {
			b.WriteString("\n" + defIndent + "┕━❯ " + insData.Definition)
		}
	}

}

// pFullExamples ,print full examples
func pFullExamples(examples []string, b *strings.Builder) {
	indent := strings.Repeat(Indent, 3)
	if len(examples) != 0 {
		for _, ex := range examples {
			b.WriteString(indent + ex + "\n")
		}
	}
}

func pExamples(examples []string, indent string, b *strings.Builder) {
	b.WriteString("\nExamples:\n")

	if len(examples) != 0 {
		for i, ex := range examples {
			b.WriteString(indent + strconv.Itoa(i+1) + ". " + ex + "\n")
		}
	}
}

// sprintSyn ,returns a string of synonyms
func (w *Word) sprintSynAnt(insType string) string {
	b := &strings.Builder{}

	opts := &pInsOpts{Type: insType, Words: true, TypeIndent: 1, WordIndent: 2}
	if w.FullDefs != nil {
		for _, fdef := range w.FullDefs {
			b.WriteString("\n" + strconv.Itoa(fdef.GroupNum) + "\n")
			for oi, ord := range fdef.Ordinals {
				pDefinition(ord, oi+1, b)
				printInstances(ord.Instances, b, opts)
			}
		}
	}

	return b.String()
}

// SprintSuggestions returns a string of suggestions
func SprintSuggestions(sugs []string) string {
	b := &strings.Builder{}

	if len(sugs) > 0 {
		b.WriteString("Did you mean?\n")
		for i, sug := range sugs {
			b.WriteString(Indent + fmt.Sprintf("%2d. ", i+1) + sug + "\n")
		}
	}

	return b.String()
}

// Arrow - ┗━❯
