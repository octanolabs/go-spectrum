package logui

import (
	"github.com/ubiq/go-ubiq/v3/log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//TODO:
//	tview:
//		- Treenode of loggers
//		- figure out if we should Limit the textview buffer at some point
//		- Keep the log messages in a capped buffer/cache, and retrieve on demand

//func MakeHandler() *log.GlogHandler {
//
//	log.H
//
//}

type LogUi struct {
	handler    *PassthroughHandler
	newLoggers chan *tview.TextView
	loggers    map[string]*tview.TextView
	uiLogger   log.Logger
}

func NewLogUi(handler *PassthroughHandler, logger log.Logger) *LogUi {
	l := &LogUi{
		handler:    handler,
		newLoggers: handler.Chan(),
		loggers:    map[string]*tview.TextView{},
		uiLogger:   logger,
	}
	return l
}

func (l *LogUi) Start() {

	app := tview.NewApplication()

	grid := tview.NewGrid().
		SetColumns(-5, -14)

	pages := tview.NewPages()

	root := tview.NewTreeNode("root_logger").
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {

		t := node.GetText()

		logger := l.loggers[t]

		if !logger.HasFocus() {
			app.SetFocus(logger)
		}

		pages.SendToFront(t)
		pages.ShowPage(t)
	})

	grid.
		AddItem(tview.NewFrame(tree), 0, 0, 1, 1, 0, 0, false).
		AddItem(pages, 0, 1, 1, 1, 0, 0, false)

	go func() {
		for {
			select {
			case logger := <-l.newLoggers:
				name := logger.GetTitle()

				l.loggers[name] = logger

				t := tview.NewTreeNode(name)
				root.AddChild(t)

				logger.SetDoneFunc(func(key tcell.Key) {
					if key == tcell.KeyEscape {
						app.SetFocus(tree)
					}
				})

				logger.SetChangedFunc(func() {

					app.Draw()
				}).
					SetBorder(true)

				frame := tview.NewFrame(logger).AddText("PRESS ESC TO RETURN TO LOGGERS; navigate log with up & down arrow keys", false, tview.AlignCenter, tcell.ColorWhite)

				pages.AddPage(name, frame, true, false)

			}
		}
	}()

	if err := app.SetRoot(grid, true).SetFocus(tree).Run(); err != nil {
		panic(err)
	}
}
