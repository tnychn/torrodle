package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/oz/osdb"
	"github.com/sirupsen/logrus"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/a1phat0ny/torrodle"
	"github.com/a1phat0ny/torrodle/client"
	"github.com/a1phat0ny/torrodle/config"
	"github.com/a1phat0ny/torrodle/models"
	"github.com/a1phat0ny/torrodle/player"
)

const version = "1.0.4"

var u, _ = user.Current()
var home = u.HomeDir
var configFile = filepath.Join(home, ".torrodle.json")
var configurations config.TorrodleConfig

var dataDir string
var subtitlesDir string

func errorPrint(arg ...interface{}) {
	c := color.New(color.FgHiRed).Add(color.Bold)
	c.Print("✘ ")
	c.Println(arg...)
}

func infoPrint(arg ...interface{}) {
	c := color.New(color.FgHiYellow)
	c.Print("[i] ")
	c.Println(arg...)
}

func pickCategory() string {
	category := ""
	prompt := &survey.Select{
		Message: "Choose a category:",
		Options: []string{"All", "Movie", "TV", "Anime", "Porn"},
	}
	survey.AskOne(prompt, &category, nil)
	return category
}

func pickProviders(options []string) []interface{} {
	chosen := []string{}
	prompt := &survey.MultiSelect{
		Message: "Choose providers:",
		Options: options,
	}
	survey.AskOne(prompt, &chosen, nil)

	providers := []interface{}{}
	for _, choice := range chosen {
		for _, provider := range torrodle.AllProviders {
			if provider.GetName() == choice {
				providers = append(providers, provider)
			}
		}
	}
	return providers
}

func inputQuery() string {
	query := ""
	prompt := &survey.Input{Message: "Search Torrents:"}
	survey.AskOne(prompt, &query, nil)
	return query
}

func pickSortBy() string {
	sortBy := ""
	prompt := &survey.Select{
		Message: "Sort by:",
		Default: "default",
		Options: []string{"default", "seeders", "leechers", "size"},
	}
	survey.AskOne(prompt, &sortBy, nil)
	return sortBy
}

func pickPlayer() string {
	options := []string{"None"}
	playerChoice := ""
	for _, p := range player.Players {
		options = append(options, p.Name)
	}
	prompt := &survey.Select{
		Message: "Player:",
		Options: options,
	}
	survey.AskOne(prompt, &playerChoice, nil)
	return playerChoice
}

func chooseResults(results []models.Source) string {
	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"#", "Name", "S", "L", "Size"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.BgHiYellowColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.BgHiGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgHiRedColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgHiCyanColor, tablewriter.FgBlackColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.FgHiCyanColor},
	)
	for i, result := range results {
		title := strings.TrimSpace(result.Title)
		isEng := utf8.RuneCountInString(title) == len(title)
		if isEng {
			if len(title) > 45 {
				title = title[:42] + "..."
			}
		} else {
			if utf8.RuneCountInString(title) > 25 {
				title = string([]rune(title)[:22]) + "..."
			}
		}
		table.Append([]string{strconv.Itoa(i + 1), title, strconv.Itoa(result.Seeders), strconv.Itoa(result.Leechers), humanize.Bytes(uint64(result.FileSize))})
	}
	table.Render()

	// Prompt choice
	choice := ""
	question := &survey.Question{
		Prompt: &survey.Input{Message: "Choice(#):"},
		Validate: func(val interface{}) error {
			index, err := strconv.Atoi(val.(string))
			if err != nil {
				return fmt.Errorf("input must be numbers")
			} else if index < 1 || index > len(results) {
				return fmt.Errorf("input range exceeded (1-%d)", len(results))
			}
			return nil
		},
	}
	survey.Ask([]*survey.Question{question}, &choice)
	return choice
}

func pickLangs() []string {
	languagesMap := map[string]string{
		"English":               "eng",
		"Chinese (traditional)": "zht",
		"Chinese (simplified)":  "chi",
		"Arabic":                "ara",
		"Hindi":                 "hin",
		"Dutch":                 "dut",
		"French":                "fre",
		"Portuguese":            "por",
		"Russian":               "rus",
	}
	languagesOpts := []string{}
	for k := range languagesMap {
		languagesOpts = append(languagesOpts, k)
	}

	chosen := []string{}
	prompt := &survey.MultiSelect{
		Message: "Choose subtitles languages:",
		Default: []string{"English"},
		Options: languagesOpts,
	}
	survey.AskOne(prompt, &chosen, nil)

	languages := []string{}
	for _, choice := range chosen {
		languages = append(languages, languagesMap[choice])
	}
	return languages
}

func chooseSubtitles(subtitles osdb.Subtitles) string {
	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"#", "Name", "Lang", "Fmt", "HI", "Size"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.BgHiYellowColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.BgHiCyanColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgHiMagentaColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgHiBlueColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.BgYellowColor, tablewriter.FgBlackColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{}, // name
		tablewriter.Colors{tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.FgHiMagentaColor},
		tablewriter.Colors{}, // hi
		tablewriter.Colors{tablewriter.FgYellowColor},
	)
	for i, sub := range subtitles {
		// hearing impaired
		var hiSymbol string
		hi, _ := strconv.Atoi(sub.SubHearingImpaired)
		if hi != 0 {
			hiSymbol = color.HiGreenString("Y")
		} else {
			hiSymbol = color.HiRedString("N")
		}
		// name
		name := strings.TrimSpace(sub.MovieReleaseName)
		if len(name) > 42 {
			name = name[:39] + "..."
		}
		// size
		size, _ := strconv.ParseUint(sub.SubSize, 10, 0)

		table.Append([]string{strconv.Itoa(i + 1), name, sub.SubLanguageID, sub.SubFormat, hiSymbol, humanize.Bytes(size)})
	}
	table.Render()

	// Prompt choice
	choice := ""
	question := &survey.Question{
		Prompt: &survey.Input{Message: "Choice(#):"},
		Validate: func(val interface{}) error {
			index, err := strconv.Atoi(val.(string))
			if err != nil {
				return fmt.Errorf("input must be numbers")
			} else if index < 1 || index > len(subtitles) {
				return fmt.Errorf("input range exceeded (1-%d)", len(subtitles))
			}
			return nil
		},
	}
	survey.Ask([]*survey.Question{question}, &choice)
	return choice
}

func getSubtitles(query string) (subtitlePath string) {
	// yes or no
	need := false
	prompt := &survey.Confirm{
		Message: "Need subtitles?",
	}
	survey.AskOne(prompt, &need, nil)
	if need == false {
		return
	}
	// pick subtitle languages
	langs := pickLangs()
	c, _ := osdb.NewClient()
	if err := c.LogIn("", "", ""); err != nil {
		errorPrint(err)
		os.Exit(1)
	}
	// search subtitles
	langstr := strings.Join(langs, ",")
	params := []interface{}{
		c.Token,
		[]struct {
			Query string `xmlrpc:"query"`
			Langs string `xmlrpc:"sublanguageid"`
		}{{
			query,
			langstr,
		}},
	}
	subtitles, err := c.SearchSubtitles(&params)
	if err != nil {
		errorPrint(err)
		os.Exit(1)
	}
	if len(subtitles) == 0 {
		errorPrint("No subtitles found")
		return
	}
	// choose subtitles
	choice := chooseSubtitles(subtitles)
	index, _ := strconv.Atoi(choice)
	subtitle := subtitles[index-1]
	// download
	fmt.Println(color.HiYellowString("[i] Downloading subtitle to"), subtitlesDir)
	subtitlePath = filepath.Join(subtitlesDir, subtitle.SubFileName)
	err = c.DownloadTo(&subtitle, subtitlePath)
	if err != nil {
		errorPrint(err)
		os.Exit(1)
	}
	// cleanup
	c.LogOut()
	c.Close()
	return
}

func startClient(player *player.Player, source models.Source, subtitlePath string) {
	// Play the video
	infoPrint("Streaming torrent...")
	// create client
	c, err := client.NewClient(dataDir, configurations.TorrentPort, configurations.HostPort)
	if err != nil {
		errorPrint(err)
		os.Exit(1)
	}
	_, err = c.SetSource(source)
	if err != nil {
		errorPrint(err)
		os.Exit(1)
	}
	// start client
	c.Start()
	// handle video playing
	if player != nil && subtitlePath != "" {
		// serve via HTTP
		c.Serve()
		fmt.Println(color.HiYellowString("[i] Serving on"), c.URL)
		// open player
		player.Start(c.URL, subtitlePath)
		fmt.Println(color.HiYellowString("[i] Launched player"), player.Name)
	}
	// handle exit signals
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func(interruptChannel chan os.Signal) {
		for range interruptChannel {
			c.Close()
			fmt.Print("\n")
			infoPrint("Exiting...")
			os.Exit(0)
		}
	}(interruptChannel)
	// print progress
	fmt.Println("File:", c.Torrent.Name())
	if player != nil {
		fmt.Println("Stream:", c.URL)
	}
	fmt.Println("Location:", filepath.Join(dataDir, c.Torrent.Name()))
	for {
		c.PrintProgress()
	}
}

func init() {
	var err error

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err = config.InitConfig(configFile)
		if err != nil {
			fmt.Printf("Error initializing config (%v): %v\n", configFile, err)
			os.Exit(1)
		}
	}

	configurations, err = config.LoadConfig(configFile)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	dataDir = configurations.DataDir
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "torrodle")
	} else if strings.HasPrefix(dataDir, "~/") {
		dataDir = filepath.Join(home, dataDir[2:]) // expand user home directoy for path in configurations file
	}
	configurations.DataDir = dataDir
	subtitlesDir = filepath.Join(dataDir, "subtitles")

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0700)
	}
	if _, err := os.Stat(subtitlesDir); os.IsNotExist(err) {
		os.Mkdir(subtitlesDir, 0700)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		DisableTimestamp:       true,
		DisableLevelTruncation: false,
	})
	logrus.SetOutput(os.Stdout)
	if configurations.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

func main() {
	name := color.HiYellowString("[torrodle v%s]", version)
	banner :=
`
	_____                          ______________     
	__  /________________________________  /__  /____ 
	_  __/  __ \_  ___/_  ___/  __ \  __  /__  /_  _ \
	/ /_ / /_/ /  /   _  /   / /_/ / /_/ / _  / /  __/
	\__/ \____//_/    /_/    \____/\__,_/  /_/  \___/

    ‣ You are using %v
`
	heart := color.HiRedString("♥︎")
	bold := color.New(color.Bold)
	// Startup
	fmt.Printf(banner, name)
	bold.Print("    Made with ")
	fmt.Print(heart)
	bold.Print(" by a1phat0ny ")
	fmt.Print("(https://github.com/a1phat0ny/torrodle)\n\n")
	logrus.Debug(configurations)

	// Stream torrent from magnet provided in command-line
	if len(os.Args) > 1 {
		// make source
		source := models.Source{
			From: "User Provided",
			Title: "Unknown",
			Magnet: os.Args[1],
		}
		// player
		playerChoice := pickPlayer()
		if playerChoice == "" {
			errorPrint("Operation aborted")
			return
		}
		var p *player.Player
		if playerChoice == "None" {
			p = nil
		} else {
			p = player.GetPlayer(playerChoice)
		}
		// start
		startClient(p, source, "")
	}

	// Prepare options and query for searching torrents
	category := pickCategory()
	if category == "" {
		errorPrint("Operation aborted")
		return
	}
	cat := torrodle.Category(strings.ToUpper(category))
	options := []string{}
	// check for availibility of each category for each provider
	for _, provider := range torrodle.AllProviders {
		if torrodle.GetCategoryURL(cat, provider.GetCategories()) != "" {
			options = append(options, provider.GetName())
		}
	}
	providers := pickProviders(options)
	if len(providers) == 0 {
		errorPrint("Operation aborted")
		return
	}
	query := inputQuery()
	query = strings.TrimSpace(query)
	if query == "" {
		errorPrint("Operation aborted")
		return
	}
	sortBy := pickSortBy()
	if sortBy == "" {
		errorPrint("Operation aborted")
		return
	}
	sb := torrodle.SortBy(strings.ToLower(sortBy))

	// Call torrodle API to search for torrents
	limit := configurations.ResultsLimit
	results := torrodle.ListResults(providers, query, limit, cat, sb)
	if len(results) == 0 {
		errorPrint("No torrents found")
		return
	}
	// Prompt for source choosing
	fmt.Print("\033c") // reset screen
	choice := chooseResults(results)
	if choice == "" {
		errorPrint("Operation aborted")
		return
	}
	index, _ := strconv.Atoi(choice)
	source := results[index-1]

	// Print source information
	fmt.Print("\033c") // reset screen
	boldYellow := color.New(color.Bold, color.FgBlue)
	boldYellow.Print("Title: ")
	fmt.Println(source.Title)
	boldYellow.Print("From: ")
	fmt.Println(source.From)
	boldYellow.Print("URL: ")
	fmt.Println(source.URL)
	boldYellow.Print("Seeders: ")
	color.Green(strconv.Itoa(source.Seeders))
	boldYellow.Print("Leechers: ")
	color.Red(strconv.Itoa(source.Leechers))
	boldYellow.Print("FileSize: ")
	color.Cyan(strconv.Itoa(int(source.FileSize)))
	boldYellow.Print("Magnet: ")
	fmt.Println(source.Magnet)

	// Player
	playerChoice := pickPlayer()
	if playerChoice == "" {
		errorPrint("Operation aborted")
		return
	}
	var p *player.Player
	var subtitlePath string
	if playerChoice == "None" {
		p = nil
	} else {
		// Get subtitles
		subtitlePath = getSubtitles(source.Title)
		p = player.GetPlayer(playerChoice)
	}

	// Start playing video...
	startClient(p, source, subtitlePath)
}
