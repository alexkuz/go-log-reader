package main

import (
	"bufio"
	"encoding/json"
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
)

type ActivePane int

const (
	ActiveTabs ActivePane = 0
	ActiveLeft ActivePane = 1
	ActiveRight ActivePane = 2
)

type LogConfig struct {
	Title string `json:"title"`
	Command string `json:"command"`
	EntryPattern string `json:"entry_pattern"`
}

type Config struct {
	Logs []LogConfig `json:"logs"`
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

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	tb.SetInputMode(tb.InputEsc)

	configPath := "go-log-reader.json"

	args := os.Args[1:]
	argsMap := map[string]string{}
	for i := 0; i < len(args); i++ {
		if (args[i] == "-c") {
			configPath = args[i+1]
			i += 1
		} else if (strings.Index(args[i], "--") == 0) {
			argsMap[args[i][2:]] = args[i+1]
			i += 1
		}
	}

  configBytes, err := os.ReadFile(configPath)
  if err != nil {
  	log.Fatalf("failed to read config: %v", err)
  }

	config := &Config{}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

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

	rightBox := customWidgets.NewRawParagraph()
	rightBox.Title = " Log View "

	debugBox := widgets.NewParagraph()
	debugBox.SetRect(0, termHeight - 4, termWidth, termHeight)	

	grid := ui.NewGrid()
	grid.SetRect(0, 2, termWidth, termHeight - 4)

	lt := ui.NewBlock()

	grid.Set(
		ui.NewRow(1.0,
			ui.NewCol(1.0/3, lt),
			ui.NewCol(2.0/3, rightBox),
		),
	)

	ui.Render(tabpane, grid, debugBox)

	logTables := []*customWidgets.RawTable{}

	for _,logConfig := range config.Logs {
		logTable := customWidgets.NewRawTable()
		logTable.SetRect(lt.Rectangle.Min.X, lt.Rectangle.Min.Y, lt.Rectangle.Max.X, lt.Rectangle.Max.Y)
		logTable.Title = fmt.Sprintf(" %s ", logConfig.Title)

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
		LogView: rightBox,
		Info: debugBox,
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
		ctx.Info.Text = e.ID
		switch e.ID {
		case "q", "<C-c>":
			quit <- true

		case "l":
			ctx.LeftHidden = !ctx.LeftHidden

			if (ctx.LeftHidden) {
				ctx.Grid.Set(
					ui.NewRow(1.0,
						ctx.LogView,
					),
				)
				ctx.LogView.BorderLeft = false
				ctx.LogView.BorderRight = false
			} else {
				ctx.Grid.Set(
					ui.NewRow(1.0,
						ui.NewCol(1.0/3, logTable),
						ui.NewCol(2.0/3, ctx.LogView),
					),
				)
				ctx.LogView.BorderLeft = true
				ctx.LogView.BorderRight = true
			}
			ui.Render(ctx.Grid)

		case "<Left>":
				ctx.Tabs.ActiveTabIndex = (ctx.Tabs.ActiveTabIndex + len(ctx.Tabs.TabNames) - 1) % len(ctx.Tabs.TabNames)
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
				logTable = ctx.LogTables[ctx.Tabs.ActiveTabIndex]
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)

		case "<Right>":
				ctx.Tabs.ActiveTabIndex = (ctx.Tabs.ActiveTabIndex + 1) % len(ctx.Tabs.TabNames)
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
				logTable = ctx.LogTables[ctx.Tabs.ActiveTabIndex]
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)

		case "<Down>":
			if ctx.ActiveRow < len(logTable.Rows) - 1 {
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
				ctx.ActiveRow += 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,ui.Color(7))
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}

		case "<Up>":
			if ctx.ActiveRow > 0 {
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
				ctx.ActiveRow -= 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,ui.Color(7))
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}

		case "<Escape>":
			if ctx.ActiveRow > -1 {
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
				ctx.ActiveRow = -1
				logTable.ActiveRowIndex = ctx.ActiveRow
				setViewText(ctx)
			}

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
		}
		ui.Render(logTable, ctx.LogView, ctx.Tabs, ctx.Info)
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
  		logTable.Rows = append([][]string{{str}}, logTable.Rows...)
  		if ctx.ActiveRow > -1 {
				logTable.RowStyles[ctx.ActiveRow] = ui.Theme.Table.Text
  			ctx.ActiveRow += 1
				logTable.RowStyles[ctx.ActiveRow] = ui.NewStyle(ui.ColorBlack,ui.Color(7))
				logTable.ActiveRowIndex = ctx.ActiveRow
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