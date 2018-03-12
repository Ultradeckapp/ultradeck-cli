package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"
	"github.com/skratchdot/open-golang/open"
	"gitlab.com/gammons/ultradeck-cli/client"
)

const (
	FrontendURL    = "https://app.ultradeck.co"
	BackendURL     = "https://api.ultradeck.co"
	DevFrontendURL = "http://localhost:3000"
	DevBackendURL  = "http://localhost:3001"
)

type Client struct {
	Conn     *client.WebsocketConnection
	ClientID string
}

func main() {
	c := &Client{ClientID: client.NewUUID()}

	if len(os.Args) == 1 {
		c.printHelpScreen()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "auth":
		c.doAuth()

	// creates a new directory wioth a deck.md in it
	// also ties it to ultradeck.co with a .ud.yml file in it
	// also initializes git repo with a .gitignore?
	case "create":
		c.authorizedCommand(c.create)

	// pushes deck (and related assets) to ultradeck.co
	// ultradeck will check timestamp, and reject if timestamp on server is newer
	// can be forced with -f
	case "push":
		c.authorizedCommand(c.push)

	// pull deck (and related assets) from ultradeck.co
	// client will check timestamps and reject if client timestamp is newer
	// must be done PER FILE
	// can be forced with -f
	case "pull":
		c.authorizedCommand(c.pull)

	// watch a directory and auto-make changes on ultradeck's server
	// uses websocket connection and other cool shit to pull this off
	case "watch":
		c.authorizedCommand(c.watch)

	// upgrade to paid
	case "upgrade":
		c.authorizedCommand(c.upgradeToPaid)

	// internal for testing
	case "check":
		c.authorizedCommand(c.checkAuth)

	// import a slide deck from ultradeck.co
	case "import":
		c.authorizedCommand(c.importDeck)
	case "present":
		c.openScreen("present")
	case "edit":
		c.openScreen("edit")
	}
}

func (c *Client) doAuth() {
	c.Conn = client.NewWebsocketConnection()
	channel := client.NewUUID()

	url := fmt.Sprintf("%s/beta-login?intermediate_token=%s", c.frontendURL(), channel)
	open.Start(url)

	c.Conn.RegisterListener(channel)

	requestChan := make(chan *client.Request)
	go c.Conn.Listen(requestChan)

	for {
		select {
		case <-c.Conn.Interrupt:
			c.debug("Interrupt")
			os.Exit(0)
			break
		case <-c.Conn.Done:
			fmt.Println("You are now authenticated!")
			os.Exit(0)
			break
		case msg := <-requestChan:
			c.processAuthResponse(msg)
		}
	}
}

func (c *Client) checkAuth(resp *client.AuthCheckResponse) {
	fmt.Printf("\nWelcome, %s! You're signed in.\n", resp.Name)
}

func (c *Client) upgradeToPaid(resp *client.AuthCheckResponse) {
	url := fmt.Sprintf("%s/auth", c.backendURL())
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("username", resp.Username)
	q.Add("name", resp.Name)
	q.Add("token", resp.Token)
	q.Add("image_url", resp.ImageUrl)
	q.Add("email", resp.Email)
	q.Add("subscription_name", resp.SubscriptionName)
	q.Add("redirect", "/account")
	req.URL.RawQuery = q.Encode()

	fmt.Printf("\nSending you to the pricing page...")
	open.Start(req.URL.String())
}

type Deck struct {
	Title string `json:"title"`
}

type CreateDeck struct {
	Deck *client.DeckConfig `json:"deck"`
}

func (c *Client) create(resp *client.AuthCheckResponse) {
	prompt := promptui.Prompt{Label: " What is the name of your deck?", Validate: c.validateInput}
	name, err := prompt.Run()
	if err != nil {
		fmt.Println("The deck needs a name!")
		os.Exit(1)
	}

	prompt2 := promptui.Prompt{Label: "Description"}
	description, _ := prompt2.Run()

	isPublic := true
	if resp.SubscriptionName != "free" {
		prompt3 := promptui.Select{Label: "Is the deck public?", Items: []string{"Yes", "No"}}
		_, isPublicResp, _ := prompt3.Run()
		if isPublicResp != "Yes" {
			isPublic = false
		}
	}

	deckConfigManager := &client.DeckConfigManager{}
	deck := deckConfigManager.NewDeck(name, description, isPublic)
	httpClient := client.NewHttpClient(resp.Token)

	createDeck := &CreateDeck{Deck: deck}
	j, _ := json.Marshal(&createDeck)
	jsonData := httpClient.PostRequest("api/v1/decks", j)

	if httpClient.Response.StatusCode == 200 {
		deckConfigManager.WriteJSON(jsonData)

		fmt.Println("Creating deck.md")
		deckConfigManager.WriteMarkdownFile("deck.md")
	} else {
		fmt.Println("Something went wrong with the request:")
		dataString := string(jsonData)
		if strings.Contains(dataString, "There is a limit") {
			fmt.Println("You can only create 1 deck with a free account.")
			fmt.Println("Run `ultradeck upgrade` to upgrade your account!")
		} else if strings.Contains(dataString, "Must be true for free plan users") {
			fmt.Println("Free accounts can only create public decks.")
			fmt.Println("Run `ultradeck upgrade` to upgrade your account!")
		} else {
			fmt.Println(dataString)
		}
	}
}

func (c *Client) validateInput(input string) error {
	if len(input) < 2 {
		return errors.New("Needs to be at least 2 characters.")
	}
	return nil
}

func (c *Client) pull(resp *client.AuthCheckResponse) {
	deckConfigManager := &client.DeckConfigManager{}
	deckConfigManager.ReadConfig()

	if !deckConfigManager.FileExists() {
		fmt.Println("Could not find deck config!")
		fmt.Println("Did you run 'ultradeck create' or 'ultradeck import' yet?")
		return
	}

	httpClient := client.NewHttpClient(resp.Token)

	url := fmt.Sprintf("api/v1/decks/%s?username=%s", deckConfigManager.GetDeckID(), resp.Username)
	jsonData := httpClient.GetRequest(url)

	if httpClient.Response.StatusCode == 200 {

		var serverDeckConfig *client.DeckConfig
		_ = json.Unmarshal(jsonData, &serverDeckConfig)

		// date on server must be equal to or greater than date on client
		if c.dateCompare(serverDeckConfig.UpdatedAt, deckConfigManager.DeckConfig.UpdatedAt) >= 0 {
			fmt.Println("Pulling changes from ultradeck.co...")
			deckConfigManager.WriteJSON(jsonData)
			deckConfigManager.WriteMarkdownFile("deck.md")

			// pull remote assets as well
			fmt.Println("Syncing assets...")
			assetManager := client.AssetManager{}
			assetManager.PullRemoteAssets(serverDeckConfig)
			fmt.Println("Done!")
		} else {
			fmt.Println("It looks like you might have local changes that are not on the server!")
			fmt.Println("Did you make changes to your deck elsewhere, or on ultradeck.co?")
			fmt.Println("You can force by running 'ultradeck pull -f'.")
		}
	} else {
		fmt.Println("Something went wrong with the request:")
		fmt.Println(string(jsonData))
	}
}

func (c *Client) push(resp *client.AuthCheckResponse) {
	deckConfigManager := &client.DeckConfigManager{}
	deckConfigManager.ReadConfig()

	if !deckConfigManager.FileExists() {
		fmt.Println("Could not find deck config!")
		fmt.Println("Did you run 'ultradeck create' yet?")
		return
	}

	fmt.Println("Pushing local changes to ultradeck.co...")

	httpClient := client.NewHttpClient(resp.Token)

	// push local assets
	assetManager := client.AssetManager{}

	// TODO:  really not sure I like this type of decorator pattern
	// can I make it cleaner?
	deckConfigManager.DeckConfig = assetManager.PushLocalAssets(resp.Token, deckConfigManager.DeckConfig)

	url := fmt.Sprintf("api/v1/decks/%s?client_id=%s", deckConfigManager.GetDeckID(), c.ClientID)
	jsonData := httpClient.PutRequest(url, deckConfigManager.PrepareJSONForUpload())

	if httpClient.Response.StatusCode == 200 {
		deckConfigManager.WriteJSON(jsonData)
		fmt.Println("Done!")
	} else {
		fmt.Println("Something went wrong with the request:")
		fmt.Println(string(jsonData))
	}
}

func (c *Client) authorizedCommand(cmd func(resp *client.AuthCheckResponse)) {
	authConfig := &client.AuthConfig{}
	if authConfig.AuthFileExists() {
		token := authConfig.GetToken()

		authCheck := &client.AuthCheck{}
		resp := authCheck.CheckAuth(token)
		resp.Token = token

		if resp.IsSignedIn {
			cmd(resp)
		} else {
			fmt.Println("\nIt does not look like you're signed in anymore.")
			fmt.Println("Please run 'ultradeck auth' to sign in again.")
		}
	} else {
		fmt.Println("\nNo auth config file found!")
		fmt.Println("Please run 'ultradeck auth' to log in.")
	}
}

func (c *Client) watch(resp *client.AuthCheckResponse) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching directory for changes...")

	done := make(chan bool)
	requestChan := make(chan *client.Request)

	c.Conn = client.NewWebsocketConnection()
	c.Conn.RegisterListener(resp.UUID)
	go c.Conn.Listen(requestChan)

	// add a listener to listen for changes to .ud.json, to push to backend
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == "./.ud.json" {
					continue
				}
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create || event.Op == fsnotify.Remove {
					c.push(resp)
				}
			case req := <-requestChan:
				// a request came in from the backend, via the websocket channel.

				// ensure the client id is not ours.  if it is, ignore. if not, do an update.
				c.debug("request ClientID = " + req.ClientID)
				c.debug("my ClientID = " + c.ClientID)
				if req.ClientID != c.ClientID {
					c.debug("No match, so initiating a pull")
					c.pull(resp)
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)

			case <-c.Conn.Interrupt:
				fmt.Println("interrupt")
				os.Exit(0)
				break
			case <-c.Conn.Done:
				fmt.Println("done")
				os.Exit(0)
				break
			}
		}
	}()

	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (c *Client) importDeck(resp *client.AuthCheckResponse) {
	httpClient := client.NewHttpClient(resp.Token)
	url := fmt.Sprintf("api/v1/decks?username=%s", resp.Username)
	jsonData := httpClient.GetRequest(url)

	type GetResponse struct {
		Decks []*client.DeckConfig
	}

	var serverDeckConfig GetResponse
	_ = json.Unmarshal(jsonData, &serverDeckConfig)

	var titles []string
	for _, deck := range serverDeckConfig.Decks {
		titles = append(titles, deck.Title)
	}

	prompt3 := promptui.Select{Label: "Which deck to import?", Items: titles}
	_, deckTitleToImport, err := prompt3.Run()
	if err != nil {
		os.Exit(1)
	}

	var selectedDeck *client.DeckConfig
	for _, deck := range serverDeckConfig.Decks {
		if deck.Title == deckTitleToImport {
			selectedDeck = deck
		}
	}

	fmt.Println("Importing deck...")

	deckConfigManager := client.NewDeckConfigManager()
	deckConfigManager.DeckConfig = selectedDeck
	deckConfigManager.WriteConfig()
	deckConfigManager.WriteMarkdownFile("deck.md")

	// pull remote assets as well
	fmt.Println("Syncing assets...")
	assetManager := client.AssetManager{}
	assetManager.PullRemoteAssets(selectedDeck)
	fmt.Println("Done!")
}

func (c *Client) openScreen(screenName string) {
	deckConfigManager := client.NewDeckConfigManager()
	deckConfigManager.ReadConfig()
	if !deckConfigManager.FileExists() {
		fmt.Println("Could not find deck config!")
		fmt.Println("Did you run 'ultradeck create' or 'ultradeck import' yet?")
	}
	authConfig := &client.AuthConfig{}
	if authConfig.AuthFileExists() {
		userData := authConfig.ReadConfig()
		shortUUID := deckConfigManager.GetDeckShortUUID()
		slug := deckConfigManager.DeckConfig.Slug
		fmt.Printf("Opening browser to %s screen...\n", screenName)
		url := fmt.Sprintf("%s/users/%s/decks/%s/%s/%s", c.frontendURL(), userData.Username, shortUUID, slug, screenName)
		open.Start(url)
	}
}

func (c *Client) printHelpScreen() {
	fmt.Println("ULTRADECK v0.1")
	fmt.Println("The ultradeck command-line utility allows you to create and manipulate decks straight from your local machine.")
	fmt.Println("When a directory is under ultradeck control, there will be a .ud.json file, a deck.md file, and any picture assets that are part of the deck.\n")

	fmt.Println("Command List for decks:")
	fmt.Println("\tcreate\t\t Create a new deck")
	fmt.Println("\timport\t\t Import a deck from ultradeck.co to the local directory")
	fmt.Println("\tpush\t\t Push local changes to ultradeck.co")
	fmt.Println("\tpull\t\t Pull remote deck changes from ultradeck.co")
	fmt.Println("\twatch\t\t Watch for changes either locally or remotely, and keep local + remote in sync")
	fmt.Println("\tpresent\t\t Open the present screen for the deck")
	fmt.Println("\tedit\t\t Open the edit screen for the deck")
	fmt.Println("\n")

	fmt.Println("Other commands:")
	fmt.Println("\tupgrade\t\t A handy link to upgrade your account")
	fmt.Println("\tcheck\t\t Check to make sure you're properly authorized with ultradeck.co.")
}

func (c *Client) dateCompare(d1 string, d2 string) int {
	t1, _ := time.Parse("2006-01-02T15:04:05.000Z", d1)
	t2, _ := time.Parse("2006-01-02T15:04:05.000Z", d2)

	if t1.Before(t2) {
		return -1
	}

	if t1.Equal(t2) {
		return 0
	}

	return 1
}

func (c *Client) processAuthResponse(req *client.Request) {
	c.debug("processAuthResponse")
	writer := client.NewAuthConfig(req.Data)
	writer.WriteAuth()
	c.Conn.CloseConnection()
}

func (c *Client) debug(msg string) {
	if 1 == 0 {
		log.Println(msg)
	}
}

func (c *Client) info(msg string) {
	log.Println(msg)
}

func (c *Client) backendURL() string {
	if os.Getenv("DEV_MODE") != "" {
		return DevBackendURL
	} else {
		return BackendURL
	}
}

func (c *Client) frontendURL() string {
	if os.Getenv("DEV_MODE") != "" {
		return DevFrontendURL
	} else {
		return FrontendURL
	}
}
