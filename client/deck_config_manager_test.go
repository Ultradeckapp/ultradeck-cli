package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMarkdown(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is slide 1
---
# here is slide 2

* with
* a
* list
`
	manager := NewDeckConfigManager()
	slides := manager.ParseMarkdown(markdown)

	assert.Equal(2, len(slides))
	assert.Equal("# Here is slide 1", slides[0].Markdown)
	assert.Equal("# here is slide 2\n\n* with\n* a\n* list", slides[1].Markdown)

	assert.Equal(1, slides[0].Position)
	assert.Equal(2, slides[1].Position)

	assert.Equal(0, slides[0].ID)
}

func TestParseMarkdownAddASlideAtBeginning(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is the new slide
---
# Here is existing slide 1
---
# Here is existing slide 2
`
	manager := &DeckConfigManager{}
	config := &DeckConfig{
		ID:          1,
		UUID:        NewUUID(),
		Title:       "Testing",
		Description: "Test",
	}

	slide1UUID := NewUUID()
	slide1 := &Slide{
		ID:             1,
		UUID:           slide1UUID,
		Position:       1,
		Markdown:       "# Here is existing slide 1",
		ColorVariation: 1,
	}

	slide2UUID := NewUUID()
	slide2 := &Slide{
		ID:             2,
		UUID:           slide2UUID,
		Position:       2,
		Markdown:       "# Here is existing slide 2",
		ColorVariation: 1,
	}

	config.Slides = append(config.Slides, slide1)
	config.Slides = append(config.Slides, slide2)

	manager.DeckConfig = config
	slides := manager.ParseMarkdown(markdown)

	assert.Equal(3, len(slides))

	assert.Equal("# Here is the new slide", slides[0].Markdown)
	assert.Equal(0, slides[0].ID)
	assert.Equal(1, slides[0].Position)

	assert.Equal("# Here is existing slide 1", slides[1].Markdown)
	assert.Equal(slide1UUID, slides[1].UUID)
	assert.Equal(2, slides[1].Position)

	assert.Equal("# Here is existing slide 2", slides[2].Markdown)
	assert.Equal(slide2UUID, slides[2].UUID)
	assert.Equal(3, slides[2].Position)
}

func TestParseMarkdownAddASlideAtMiddle(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is existing slide 1
---
# Here is the new slide
---
# Here is existing slide 2
`
	manager := &DeckConfigManager{}
	config := &DeckConfig{
		ID:          1,
		Title:       "Testing",
		Description: "Test",
	}

	slide1 := &Slide{
		ID:             1,
		UUID:           NewUUID(),
		Position:       1,
		Markdown:       "# Here is existing slide 1",
		ColorVariation: 1,
	}

	slide2 := &Slide{
		ID:             2,
		UUID:           NewUUID(),
		Position:       2,
		Markdown:       "# Here is existing slide 2",
		ColorVariation: 1,
	}

	config.Slides = append(config.Slides, slide1)
	config.Slides = append(config.Slides, slide2)

	manager.DeckConfig = config
	slides := manager.ParseMarkdown(markdown)

	assert.Equal(3, len(slides))

	assert.Equal("# Here is existing slide 1", slides[0].Markdown)
	assert.Equal(1, slides[0].ID)
	assert.Equal(1, slides[0].Position)

	assert.Equal("# Here is the new slide", slides[1].Markdown)
	assert.Equal(2, slides[1].Position)

	assert.Equal("# Here is existing slide 2", slides[2].Markdown)
	assert.Equal(2, slides[2].ID)
	assert.Equal(3, slides[2].Position)
}

func TestParseMarkdownRemoveASlideAtMiddle(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is existing slide 1
---
# Here is existing slide 3
`
	manager := &DeckConfigManager{}
	config := &DeckConfig{
		ID:          1,
		Title:       "Testing",
		Description: "Test",
	}

	slide1 := &Slide{
		ID:             1,
		UUID:           NewUUID(),
		Position:       1,
		Markdown:       "# Here is existing slide 1",
		ColorVariation: 1,
	}

	slide2 := &Slide{
		ID:             2,
		UUID:           NewUUID(),
		Position:       2,
		Markdown:       "# Here is existing slide 2",
		ColorVariation: 1,
	}

	slide3 := &Slide{
		ID:             3,
		UUID:           NewUUID(),
		Position:       3,
		Markdown:       "# Here is existing slide 3",
		ColorVariation: 1,
	}

	config.Slides = append(config.Slides, slide1)
	config.Slides = append(config.Slides, slide2)
	config.Slides = append(config.Slides, slide3)

	manager.DeckConfig = config
	slides := manager.ParseMarkdown(markdown)

	assert.Equal(2, len(slides))

	assert.Equal("# Here is existing slide 1", slides[0].Markdown)
	assert.Equal(1, slides[0].Position)
	assert.Equal(1, slides[0].ID)

	assert.Equal("# Here is existing slide 3", slides[1].Markdown)
	assert.Equal(3, slides[1].ID)
	assert.Equal(2, slides[1].Position)
}

func TestParseMarkdownAddASlideAtEnd(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is yet another great slide
---
# And another one
`
	manager := &DeckConfigManager{}
	config := &DeckConfig{
		ID:          1,
		Title:       "Testing",
		Description: "Test",
	}

	slide := &Slide{
		ID:             1,
		UUID:           NewUUID(),
		Position:       1,
		Markdown:       "# Here is yet another great slide",
		ColorVariation: 1,
	}

	config.Slides = append(config.Slides, slide)

	manager.DeckConfig = config

	slides := manager.ParseMarkdown(markdown)

	assert.Equal(2, len(slides))
	assert.Equal("# Here is yet another great slide", slides[0].Markdown)
	assert.Equal("# And another one", slides[1].Markdown)

	assert.Equal(1, slides[0].ID)
	assert.Equal(1, slides[0].Position)

	assert.Equal(0, slides[1].ID)
	assert.Equal(2, slides[1].Position)
}

func TestNewDeck(t *testing.T) {
	assert := assert.New(t)

	manager := &DeckConfigManager{}
	deck := manager.NewDeck("test title", "test description")

	assert.Equal("test title", deck.Title)
	assert.Equal("test description", deck.Description)
	assert.Equal(1, len(deck.Slides))

	assert.NotEqual(nil, deck.Slides[0].UUID)
	assert.Equal("# New Slide", deck.Slides[0].Markdown)
}
