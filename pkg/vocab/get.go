package vocab

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var programName = "mpg123"

// Get fetches word from vocabulary.com
// and returns word definition, and suggestions if any
func Get(word string) (*Word, []string) {
	var doc *goquery.Document
	var err error

	wordDef := &Word{Word: word}
	wordURL := siteBaseURL + word
	wordDef.wg = &sync.WaitGroup{}

	var sugs []string

	// we try to fetch document and examples asynchronously,
	// as soon as document is fetched, we try to fetch audio asynchronously,
	// then we process the doc
	wordDef.wg.Add(1)
	go func() {
		doc, err = goquery.NewDocument(wordURL)
		if err != nil {
			log.Fatalln(err)
		}

		sugs = suggestions(word, doc)
		if len(sugs) != 0 {
			wordDef.wg.Done()
			return
		}

		// fetch examples
		wordDef.wg.Add(1)
		go fetchExamples(wordDef, word)

		// fetch audio asynchronously
		wordDef.wg.Add(1)
		go fetchAudio(wordDef, doc)

		// process doc
		shortAndLong(wordDef, doc)
		primaryAndFull(wordDef, doc)
		wordDef.wg.Done()
	}()

	// wait
	wordDef.wg.Wait()

	if len(sugs) > 0 {
		return nil, sugs
	}

	return wordDef, nil
}

// suggestions are given in case:
//	- Spelling Mistake
//  - Same word is captalized
func suggestions(word string, doc *goquery.Document) (suggestions []string) {
	doc.Find(".suggestions").Find("span.word").Each(func(i int, sel *goquery.Selection) {
		suggestions = append(suggestions, sel.Text())
	})

	return
}

func shortAndLong(word *Word, doc *goquery.Document) {
	// short definition
	word.Short = doc.Find(".short").Text()

	// long definition
	word.Long = doc.Find(".long").Text()
}

func primaryAndFull(word *Word, doc *goquery.Document) {
	doc.Find(".definitions").Find(".section.definition").Each(func(i int, sel *goquery.Selection) {

		// it's not possible to decide, if this section has
		// primary defs or section defs
		// following functions will fill data as required
		primary(word, sel)
		fullDefs(word, sel)
	})
}

func primary(word *Word, sel *goquery.Selection) {

	// primary definitions are in tbody tag
	sel.Find("tbody").Find("a").Each(func(i int, sel *goquery.Selection) {
		if val, ok := sel.Attr("href"); ok {
			word.PrimaryIDs = append(word.PrimaryIDs, val[1:]) // href="#s104789"
		}
	})
}

func fullDefs(word *Word, sel *goquery.Selection) {

	// full defs or defs have group class
	sel.Find(".group").Each(func(i int, sel *goquery.Selection) {
		word.FullDefs = append(word.FullDefs, *fullDef(sel))
	})
}

func fullDef(sel *goquery.Selection) *FullDef {
	fullDef := &FullDef{}

	groupNum := sel.Find(".groupNumber").Text()

	fullDef.GroupNum, _ = strconv.Atoi(groupNum)

	sel.Find("div").Each(func(i int, sel *goquery.Selection) {
		if id, ok := sel.Attr("id"); ok {
			fullDef.Ordinals = append(fullDef.Ordinals, *ordinal(id, sel))
		}
	})

	return fullDef
}

func ordinal(id string, sel *goquery.Selection) *Ordinal {
	def := &Ordinal{}

	defTree := sel.Find("h3.definition")
	defData := defTree.Nodes[0].LastChild.Data // https://github.com/PuerkitoBio/goquery/issues/213#issuecomment-491603678

	def.ID = id
	def.ClassType, _ = defTree.Find("a").Attr("title") // noun, adjective, ...
	def.Definition = strings.TrimSpace(defData)

	// examples for particular definition
	sel.Find("div.example").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		text = strings.ReplaceAll(text, "\n", "")
		def.Examples = append(def.Examples, text)
	})

	sel.Find("dl.instances").Each(func(i int, sel *goquery.Selection) {
		def.Instances = append(def.Instances, *instance(sel))
	})

	return def
}

// instance contains synonms or type or type of or ...
func instance(sel *goquery.Selection) *Instance {
	instance := &Instance{}

	typeName := sel.Find("dt").Text()
	if len(typeName) > 1 {
		typeName = typeName[:len(typeName)-1]
	}

	instance.Type = typeName

	sel.Find("dd").Each(func(i int, sel *goquery.Selection) {

		iData := &InstanceData{}

		sel.Find("a.word").Each(func(i int, sel *goquery.Selection) {
			iData.Words = append(iData.Words, sel.Text())
		})

		iData.Definition = sel.Find("div.definition").Text()

		if !iData.isEmpty() {
			instance.Datas = append(instance.Datas, *iData)
		}
	})

	return instance
}

func (iData *InstanceData) isEmpty() bool {
	return iData.Definition == "" && len(iData.Words) == 0
}

// fetchAudio if available from definition's response
func fetchAudio(word *Word, doc *goquery.Document) {

	doc.Find(".audio").Each(func(i int, sel *goquery.Selection) {

		if key, ok := sel.Attr("data-audio"); ok {
			audioURL := audioBaseURL + key + ".mp3"

			resp, err := http.Get(audioURL)
			if err != nil {
				log.Println(err)
			}

			audio, _ := ioutil.ReadAll(resp.Body)

			word.Audios = append(word.Audios, audio)
		}
	})
	word.wg.Done()
}

// fetchExamples from api for `wordStr`
func fetchExamples(word *Word, wordStr string) {

	exURL := exBaseURL + wordStr + "&maxResults=7"

	resp, err := http.Get(exURL)
	if err != nil {
		log.Fatalln("resp", err)
	}

	var examples List

	// decode
	json.NewDecoder(resp.Body).Decode(&examples)

	if !(len(examples.Result.Sentences) < 1) {
		for _, v := range examples.Result.Sentences {
			word.Examples = append(word.Examples, v.Sentence)
		}
	}

	word.wg.Done()
}

// PlayAudio ,plays audio
func (w *Word) PlayAudio() {
	for _, audio := range w.Audios {
		cmd := exec.Command(programName, "-")
		cmd.Stdin = bytes.NewReader(audio)
		cmd.Run()
	}
}
