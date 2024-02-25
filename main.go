package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	customWidgets "replika.com/log-reader/widgets"
	tb "github.com/nsf/termbox-go"
	"github.com/spf13/viper"
)

type ActivePane int

const (
	ActiveTabs ActivePane = 0
	ActiveLeft ActivePane = 1
	ActiveRight ActivePane = 2
)

type LogConfig struct {
	Title string `mapstructure:"title"`
	Command string `mapstructure:"command"`
	EntryPattern string `mapstructure:"entry_pattern"`
}

type Config struct {
	Logs []LogConfig `mapstructure:"logs"`
}

type Context struct {
	ActivePane ActivePane
	ActiveRow int
	Config *Config

	Grid *ui.Grid
	Tabs *widgets.TabPane
	LogTables []*customWidgets.RawTable
	LogView *customWidgets.RawParagraph
	Info *widgets.Paragraph
	LogTableCell ui.GridItem
	LeftHidden bool
	RightHidden bool
}

const rowSeparatorColor = ui.Color(240)
const selectedRowColor = ui.Color(7)

func main() {
	viper.SetConfigName(".go-log-reader")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME") 

	args := os.Args[1:]
	argsMap := map[string]string{}
	for i := 0; i < len(args); i++ {
		if (args[i] == "-c") {
			// configPath = args[i+1]
			viper.SetConfigFile(args[i+1])
			i += 1
		} else if (strings.Index(args[i], "--") == 0) {
			argsMap[args[i][2:]] = args[i+1]
			i += 1
		}
	}

	config := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("%v", err)
		} else {
			log.Fatalf("%v", err)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	tb.SetInputMode(tb.InputEsc)

	for key, val := range argsMap {
		for j := 0; j < len(config.Logs); j++ {
			config.Logs[j].Command = strings.Replace(config.Logs[j].Command, fmt.Sprintf("${%s}", key), val, -1)
		}
	}

	termWidth, termHeight := ui.TerminalDimensions()

	tabNames := []string{}

	for _,logConfig := range config.Logs {
		tabNames = append(tabNames, logConfig.Title)
	}

	tabpane := widgets.NewTabPane(tabNames...)
	tabpane.SetRect(0, 1, termWidth, 2)
	tabpane.Border = false
	tabpane.ActiveTabStyle.Fg = ui.ColorCyan
	tabpane.ActiveTabStyle.Modifier = ui.ModifierBold | ui.ModifierUnderline
	tabpane.InactiveTabStyle.Modifier = ui.ModifierBold

	logView := customWidgets.NewRawParagraph()
	logView.PaddingLeft = 1
	logView.Title = " Log Entry "

	info := widgets.NewParagraph()
	info.PaddingLeft = 1
	info.PaddingRight = 1
	info.SetRect(0, termHeight - 4, termWidth, termHeight)
	info.Title = " Info "
	info.Text = "Press [l](fg:yellow) to show/hide log list"

	grid := ui.NewGrid()
	grid.SetRect(0, 2, termWidth, termHeight - 4)

	lt := ui.NewBlock()

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(1.0/3, lt),
			ui.NewCol(2.0/3, logView),
		),
	)

	ui.Render(tabpane, grid, info)

	logTables := []*customWidgets.RawTable{}

	for range config.Logs {
		logTable := customWidgets.NewRawTable()
		logTable.Border = false
		logTable.SeparatorStyle = ui.NewStyle(rowSeparatorColor)
		logTable.SetRect(lt.Rectangle.Min.X, lt.Rectangle.Min.Y, lt.Rectangle.Max.X, lt.Rectangle.Max.Y)
		// logTable.Title = fmt.Sprintf(" %s ", logConfig.Title)

		logTable.ColumnWidths = []int{termWidth / 2}
		logTables = append(logTables, logTable)
	}

	ui.Render(logTables[0])

	ctx := &Context{
		ActivePane: ActiveTabs,
		ActiveRow: -1,
		Config: config,
		Tabs: tabpane,
		LogTables: logTables,
		LogView: logView,
		Info: info,
		Grid: grid,
		LeftHidden: false,
		RightHidden: false,
	}

	for i := range config.Logs {
		go listenLog(ctx, i)
	}

	quit := make(chan bool, 1)

	go listenKeys(ctx, quit)

	<-quit
}

func listenKeys(ctx *Context, quit chan bool) {
	uiEvents := ui.PollEvents()
	logTable := ctx.LogTables[ctx.Tabs.ActiveTabIndex]

	for {
		e := <-uiEvents
		// ctx.Info.Text = e.ID
		
		switch e.ID {
		case "q", "<C-c>":
			quit <- true

		case "l":
			ctx.LeftHidden = !ctx.LeftHidden
			updateGridLayout(ctx)
			ui.Render(ctx.Grid, logTable, ctx.LogView)

		case "<Left>":
				ctx.Tabs.ActiveTabIndex = (ctx.Tabs.ActiveTabIndex + len(ctx.Tabs.TabNames) - 1) % len(ctx.Tabs.TabNames)
				resetRowStyles(ctx)
				logTable = ctx.LogTables[ctx.Tabs.ActiveTabIndex]
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
				ui.Render(logTable, ctx.LogView, ctx.Tabs)

		case "<Right>":
				ctx.Tabs.ActiveTabIndex = (ctx.Tabs.ActiveTabIndex + 1) % len(ctx.Tabs.TabNames)
				resetRowStyles(ctx)
				logTable = ctx.LogTables[ctx.Tabs.ActiveTabIndex]
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
				ui.Render(logTable, ctx.LogView, ctx.Tabs)

		case "<Down>":
			if ctx.ActiveRow < len(logTable.Rows) - 1 {
				resetRowStyles(ctx)
				ctx.ActiveRow += 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,selectedRowColor)
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}
			ui.Render(logTable, ctx.LogView)

		case "<Up>":
			if ctx.ActiveRow > 0 {
				resetRowStyles(ctx)
				ctx.ActiveRow -= 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,selectedRowColor)
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}
			ui.Render(logTable, ctx.LogView)

		case "<Escape>":
			if ctx.ActiveRow > -1 {
				resetRowStyles(ctx)
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}
			ui.Render(logTable, ctx.LogView)

		case "<Tab>":
			ctx.ActivePane = (ctx.ActivePane +1) % 3

			if (ctx.ActivePane == ActiveTabs) {
				ctx.Tabs.ActiveTabStyle.Modifier = ui.ModifierBold | ui.ModifierUnderline
				ctx.Tabs.InactiveTabStyle.Modifier = ui.ModifierBold
			} else {
				ctx.Tabs.ActiveTabStyle.Modifier = ui.ModifierClear
				ctx.Tabs.InactiveTabStyle.Modifier = ui.ModifierClear				
			}

			if (ctx.ActivePane == ActiveLeft) {
				logTable.TitleStyle.Modifier = ui.ModifierBold
				logTable.BorderStyle.Modifier = ui.ModifierBold
			} else {
				logTable.TitleStyle.Modifier = ui.ModifierClear
				logTable.BorderStyle.Modifier = ui.ModifierClear
			}

			if (ctx.ActivePane == ActiveRight) {
				ctx.LogView.TitleStyle.Modifier = ui.ModifierBold
				ctx.LogView.BorderStyle.Modifier = ui.ModifierBold
			} else {
				ctx.LogView.TitleStyle.Modifier = ui.ModifierClear
				ctx.LogView.BorderStyle.Modifier = ui.ModifierClear
			}
			ui.Render(logTable, ctx.LogView, ctx.Tabs, ctx.Info)

		case "<Resize>":
			termWidth, termHeight := ui.TerminalDimensions()
			ctx.Tabs.SetRect(0, 1, termWidth, 2)
			ctx.Info.SetRect(0, termHeight - 4, termWidth, termHeight)
			ctx.Grid.SetRect(0, 2, termWidth, termHeight - 4)
			updateGridLayout(ctx)
			ui.Render(ctx.Grid, logTable, ctx.LogView, ctx.Tabs, ctx.Info)
		}
	}
}

func resetRowStyles(ctx *Context) {
	logTable := ctx.LogTables[ctx.Tabs.ActiveTabIndex]
	// resetRowStyles(ctx)
	logTable.RowStyles = map[int]ui.Style{}
}

func updateGridLayout(ctx *Context) {
	logTable := ctx.LogTables[ctx.Tabs.ActiveTabIndex]
	if (ctx.LeftHidden) {
		ctx.Grid.Set(
			ui.NewRow(1.0,
				ctx.LogView,
			),
		)
		ctx.LogView.Border = false
		ctx.LogView.Title = ""
	} else {
		ctx.Grid.Set(
			ui.NewRow(1.0,
				ui.NewCol(1.0/3, logTable),
				ui.NewCol(2.0/3, ctx.LogView),
			),
		)
		ctx.LogView.Border = true
		ctx.LogView.Title = " Log View "
	}
}

func setViewText(ctx *Context) {
	logTable := ctx.LogTables[ctx.Tabs.ActiveTabIndex]
	row := ctx.ActiveRow
	if row == -1 {
		row = 0
	}

	if (len(logTable.Rows) > row) {
		ctx.LogView.Text = logTable.Rows[row][0]
	} else {
		ctx.LogView.Text = ""
	}
}

func listenLog(ctx *Context, index int) {
	cmdArr := strings.Split(ctx.Config.Logs[index].Command, " ")
  cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	logStartRe, _ := regexp.Compile(ctx.Config.Logs[index].EntryPattern)

  logTable := ctx.LogTables[index]

  stdout, _ := cmd.StdoutPipe()
  cmd.Start()

  scanner := bufio.NewScanner(stdout)

  for scanner.Scan() {
    str := scanner.Text()
    if logStartRe.MatchString(str) {
  		logTable.Rows = append([][]string{{strings.TrimSpace(str)}}, logTable.Rows...)
  		if ctx.ActiveRow > -1 {
				logTable.RowStyles = map[int]ui.Style{}
  			ctx.ActiveRow += 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,selectedRowColor)
				logTable.ActiveRowIndex = ctx.ActiveRow
				if (ctx.Tabs.ActiveTabIndex == index) {
					ui.Render(logTable)
				}
  		}
    } else {
    	if len(logTable.Rows) > 0 {
    		logTable.Rows[0][0] += "\n" + str
    	}
    }

		if (ctx.Tabs.ActiveTabIndex == index) {
	    if len(logTable.Rows) > 0 {
	    	row := 0
	    	if ctx.ActiveRow > -1 {
	    		row = ctx.ActiveRow
	    	}
	    	ctx.LogView.Text = logTable.Rows[row][0]
	    }

			ui.Render(logTable, ctx.LogView)
		}	
  }
  cmd.Wait()
}