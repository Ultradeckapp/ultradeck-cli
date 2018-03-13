This is the command-line utility for [Ultradeck](https://ultradeck.co).  `ultradeck` allows you to create and manipulate decks straight from your local machine in a seamless way.

## Installation

There are a couple of options to install:

* If using homebrew, Ultradeck is available via `brew install ultradeck`.
* You can download the binaries via the [Releases](https://github.com/gammons/ultradeck-cli/releases) page on Github.
* If you're using Go, you can also run `go install github.com/gammons/ultradeck-cli` to install the binary for your system.

## Quickstart tutorial

#### 1. Authenticate yourself.

You'll need to authenticate yourself the first time you use `ultradeck`.

1. Run `ultradeck auth`
2. You'll be taken to ultradeck's login page, where you'll login via Github or Google.
3. Once done, head back to your terminal. Ultradeck will have completed the auth and will have written a file at `~/.config/ultradeck/auth.json`, which includes your authentication token.

#### 2. Create a new deck using the cli.

Run `ultradeck create` in a new directory.  Ultradeck will ask you your deck's title, description, and visibility setting.

`ultradeck` will create 2 files in your directory:

1. deck.md <- this is your deck.
2. .ud.json <- this is the internal file that keeps your deck in sync with what is on the [ultradeck.co](https://ultradeck.co) webapp.

#### 3. Open up the ultradeck editor webapp

In your directory, run `ultradeck edit`.  This will bring up the editor screen on [ultradeck.co](https://ultradeck.co).

The editor screen is what you'll use to get a feel for how your deck looks, change colors + themes, and write presenter notes.

#### 4. In a separate terminal, head to your directory and run `ultradeck watch`.

The `watch` command will watch both [ultradeck.co](https://ultradeck.co) and `deck.md` for changes, and bi-directionally update.

Using your favorite editor, open `deck.md` and add a few slides to your deck:

```markdown
## Slide 1

* Here's the first bullet

---

## Slide 2

* And here is another item
```

When you write the file, `ultradeck` will automatically push your changes to [ultradeck.co](https://ultradeck.co).  This will allow you to quickly iterate on your deck and get the general idea across.

## Command reference

**Authentication**

* `auth`:  Authenticate with ultradeck.  Once authenticated, you can use the `ultradeck` command in any directory without the need to re-authenticate each time.

**Create and import decks**

* `create`: Create a new deck
* `import`: Import a deck from [ultradeck.co](https://ultradeck.co) to the current directory

**Pushing and pulling changes**

* `push`: push local changes to [ultradeck.co](https://ultradeck.co)
* `pull`: pull remote deck changes from [ultradeck.co](https://ultradeck.co) (be sure to reload deck.md in your editor!)
* `watch`: Watch for changes locally and remotely, and keep local + remote in sync

**Opening pages on ultradeck.co**

* `present`: Show the deck view on [ultradeck.co](https://ultradeck.co) for the current deck
* `edit`: Show the edit deck view on [ultradeck.co](https://ultradeck.co) for the current deck

**Other**

* `check`: check to ensure you're correctly logged in
* `upgrade`: Go to the pricing page to upgrade your account

## Managing images and other assets

Any images in the same directory as `deck.md` will be treated as [assets](https://docs.ultradeck.co/#assets) for the deck.  Assets are available for use in your slides.

Example:

```
➜  tree -a
.
├── deck.md
├── porsche.jpg
└── .ud.json
```

You can reference the image in your deck like so:

```markdown
![](porsche.jpg)

# Cool cars!
```

When you run `ultradeck push`, `porsche.jpg` will be uploaded to ultradeck.co as an asset.

## Tips for using Git with an ultradeck directory

You're encouraged to put `deck.md`, any assets, _and_ `.ud.json` under git control.

## The ud.json file

`ud.json` is a representation of your deck from ultradeck.co.  It includes information about slide themes, colors, presenter notes, custom CSS, etc.  When you run `ultradeck pull`, all information about the deck is synced over to `.ud.json`.

This is why it's important to have `.ud.json` under git control, so you can always revert back changes to the actual presentation attributes of your deck, if you need to.

## The deck.md file

`ultradeck` always looks for a file called `deck.md` which is the Markdown representation of your deck.  There is no difference in the markdown needed to create decks.

Slides in `deck.md` will be parsed with the [horizontal rule](https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet#hr) markdown command, specifically the 3 dashes: `---`.

Example:

```markdown
# Slide 1

---

# Slide 2

---

# Slide 3
```

## Keeping local decks in sync with ultradeck.co

`ultradeck` does not support fancy merge algorithms.  It will simply copy whats on remote to local if you're running `ultradeck pull`, and will blindly push what's local to remote if you're running `ultradeck push`. This is why it's important to keep local decks under version control and being careful about the changes you make on ultradeck.co.

## Importing a file from Deckset

Ultradeck has some support for markdown files created for use with [Deckset](https://www.decksetapp.com/).  To use a Deckset file with ultradeck, do the following:

1. In a new directory, create a new deck with `ultradeck create`
2. copy the deckset markdown file to the directory, and name it `deck.md`
3. run `ultradeck push`.
