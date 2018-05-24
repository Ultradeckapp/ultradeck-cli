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
	UUID        string   `json:"uuid"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	UpdatedAt   string   `json:"updated_at"`
	Slides      []*Slide `json:"slides_attributes"`
	Assets      []*Asset `json:"assets_attributes"`
}

type Slide struct {
	ID             int    `json:"id"`
	UUID           string `json:"uuid"`
	Position       int    `json:"position"`
	Markdown       string `json:"markdown"`
	PresenterNotes string `json:"presenter_notes"`
	ThemeName      string `json:"theme_name"`
	Layout         string `json:"layout"`
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

func (d *DeckConfigManager) NewDeck(title string, description string) *DeckConfig {
	deck := &DeckConfig{
		UUID:        NewUUID(),
		Title:       title,
		Description: description,
		Assets:      []*Asset{},
	}

	slide := &Slide{
		Position:       1,
		UUID:           NewUUID(),
		Markdown:       "# New Slide",
		ColorVariation: 1,
		ThemeName:      "bebas",
	}

	deck.Slides = append(deck.Slides, slide)
	return deck
}

// Write JSON coming back from backend.
// Used in creating a new deck and writing .ud.json for the first time.
func (d *DeckConfigManager) WriteJSON(jsonData []byte) {
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
	d.DeckConfig.Slides = d.ParseDeckMDFile()

	deck := &Deck{Config: d.DeckConfig}

	j, _ := json.Marshal(&deck)

	return j
}

func (d *DeckConfigManager) GetDeckID() string {
	return d.DeckConfig.UUID
}

func (d *DeckConfigManager) GetDeckShortUUID() string {
	return d.DeckConfig.UUID[0:13]
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
	var usedSlides []string

	for i, markdown := range splitted {
		// attempt to find the previous slide from the deckConfig
		var slideFromConfig *Slide
		var firstSlide *Slide

		if d.DeckConfig != nil {
			firstSlide = d.DeckConfig.Slides[0]

			for i := range d.DeckConfig.Slides {
				if d.DeckConfig.Slides[i].Markdown == strings.TrimSpace(markdown) && !contains(usedSlides, d.DeckConfig.Slides[i].UUID) {
					slideFromConfig = d.DeckConfig.Slides[i]
					usedSlides = append(usedSlides, slideFromConfig.UUID)
				}
			}
		}

		newSlide := &Slide{
			Markdown: strings.TrimSpace(markdown),
			Position: (i + 1),
		}

		if slideFromConfig != nil {
			newSlide.ID = slideFromConfig.ID
			newSlide.UUID = slideFromConfig.UUID
			newSlide.PresenterNotes = slideFromConfig.PresenterNotes
			newSlide.ThemeName = slideFromConfig.ThemeName
			newSlide.Layout = slideFromConfig.Layout
			newSlide.ColorVariation = slideFromConfig.ColorVariation
		} else if firstSlide != nil {
			newSlide.UUID = NewUUID()
			newSlide.ThemeName = firstSlide.ThemeName
			newSlide.ColorVariation = firstSlide.ColorVariation
		} else {
			// sane defaults.
			newSlide.UUID = NewUUID()
			newSlide.ThemeName = "bebas"
			newSlide.ColorVariation = 1
		}

		slides = append(slides, newSlide)
	}
	return slides
}

func (d *DeckConfigManager) WriteMarkdownFile(filename string) {
	markdown := ""

	for i, slide := range d.DeckConfig.Slides {
		if i > 0 {
			markdown += "\n\n---\n\n"
		}
		markdown += slide.Markdown
	}

	// read the current deck.md file and see if it needs updating
	currentMarkdown, _ := ioutil.ReadFile("deck.md")
	currentMarkdownString := string(currentMarkdown[:])
	if strings.TrimSpace(currentMarkdownString) == strings.TrimSpace(markdown) {
		return
	}
	if err := ioutil.WriteFile(filename, []byte(markdown), 0644); err != nil {
		log.Println("Error writing deck.md: ", err)
	}
}

func (d *DeckConfigManager) FileExists() bool {
	if _, err := os.Stat(".ud.json"); os.IsNotExist(err) {
		return false
	}
	return true
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
