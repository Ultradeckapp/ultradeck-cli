package client

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Deck struct {
	Config *DeckConfig `json:"deck"`
}

type DeckConfig struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	IsPublic    bool     `json:"is_public"`
	ThemeID     int      `json:"theme_id"`
	UpdatedAt   string   `json:"updated_at"`
	Slides      []*Slide `json:"slides_attributes"`
	Assets      []*Asset `json:"assets_attributes"`
}

type Slide struct {
	ID             int    `json:"id"`
	Position       int    `json:"position"`
	Markdown       string `json:"markdown"`
	PresenterNotes string `json:"presenter_notes"`
	ColorVariation int    `json:"color_variation"`
}

type Asset struct {
	ID        int    `json:"id"`
	Filename  string `json:"filename"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updated_at"`
}

type DeckConfigManager struct {
	DeckConfig *DeckConfig
}

func NewDeckConfigManager() *DeckConfigManager {
	manager := &DeckConfigManager{}
	manager.ReadConfig()
	return manager
}

// Write JSON coming back from backend.
// Used in creating a new deck and writing .ud.json for the first time.
func (d *DeckConfigManager) Write(jsonData []byte) {
	var deckConfig *DeckConfig
	if err := json.Unmarshal(jsonData, &deckConfig); err != nil {
		log.Println("Error writing deck", err)
	}

	d.DeckConfig = deckConfig
	d.WriteConfig()
}

// lower-level function to write the DeckConfig to .ud.json
func (d *DeckConfigManager) WriteConfig() {
	marshalledData, _ := json.Marshal(d.DeckConfig)
	if err := ioutil.WriteFile(".ud.json", marshalledData, 0644); err != nil {
		log.Println("Error writing deck config: ", err)
	}
}

// read .ud.json and store data in DeckConfig struct
func (d *DeckConfigManager) ReadConfig() {
	if !d.FileExists() {
		return
	}

	data, err := ioutil.ReadFile(".ud.json")
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	var deckConfig *DeckConfig
	err = json.Unmarshal(data, &deckConfig)
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	d.DeckConfig = deckConfig
}

// prepares what's stored in deckConfig to be uploaded to server
func (d *DeckConfigManager) PrepareJSONForUpload() []byte {
	d.ReadConfig()

	d.DeckConfig.Slides = d.ParseDeckMDFile()

	deck := &Deck{Config: d.DeckConfig}

	j, _ := json.Marshal(&deck)

	return j
}

func (d *DeckConfigManager) GetDeckID() int {
	d.ReadConfig()
	return d.DeckConfig.ID
}

// reads the markdown from deck.md file and returns a slide array of slides
func (d *DeckConfigManager) ParseDeckMDFile() []*Slide {
	markdown, err := ioutil.ReadFile("deck.md")
	if err != nil {
		log.Println("I'm expecting your markdown file to be named deck.md, but I couldn't read it!: ", err)
	}
	return d.ParseMarkdown(string(markdown[:]))
}

func (d *DeckConfigManager) ParseMarkdown(markdown string) []*Slide {
	splitted := strings.Split(string(markdown), "---\n")
	var slides []*Slide

	for i, markdown := range splitted {
		// attempt to find the previous slide from the deckConfig
		var previousSlide *Slide

		if d.DeckConfig != nil {
			for i := range d.DeckConfig.Slides {
				if d.DeckConfig.Slides[i].Markdown == strings.TrimSpace(markdown) {
					previousSlide = d.DeckConfig.Slides[i]
				}
			}
		}

		newSlide := &Slide{Position: (i + 1), Markdown: strings.TrimSpace(markdown)}

		if previousSlide != nil {
			newSlide.ID = previousSlide.ID
			newSlide.PresenterNotes = previousSlide.PresenterNotes
			newSlide.ColorVariation = previousSlide.ColorVariation
		}

		slides = append(slides, newSlide)
	}
	return slides
}

func (d *DeckConfigManager) FileExists() bool {
	if _, err := os.Stat(".ud.json"); os.IsNotExist(err) {
		return false
	}
	return true
}
