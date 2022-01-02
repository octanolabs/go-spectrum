package logui

import (
	"github.com/rivo/tview"

	"github.com/ubiq/go-ubiq/v6/log"
)

type PassthroughHandler struct {
	seenModules []LoggerContext
	handler     splitterHandler
	ch          chan *tview.TextView
}

func NewPassThroughHandler(ch chan *tview.TextView) *PassthroughHandler {
	return &PassthroughHandler{
		seenModules: nil,
		handler:     splitterHandler{},
		ch:          ch,
	}
}

func (h *PassthroughHandler) Chan() chan *tview.TextView {
	return h.ch
}

func (h *PassthroughHandler) Log(r *log.Record) error {

	ctx := getContext(r.Ctx)

	for _, v := range h.seenModules {
		if v.Equals(ctx) {
			return h.handler.Log(r)
		}
	}

	h.seenModules = append(h.seenModules, ctx)
	newWindows := h.handler.UpdateRecords(h.seenModules)

	if len(newWindows) > 0 {
		for _, v := range newWindows {
			h.ch <- v
		}
	}

	return h.handler.Log(r)
}

type splitterHandler struct {
	filters        []LoggerContext
	filterHandlers []log.Handler
	multiHandler   log.Handler
}

func (h *splitterHandler) Log(r *log.Record) error {
	return h.multiHandler.Log(r)
}

func (h *splitterHandler) UpdateRecords(filters []LoggerContext) (newWindows []*tview.TextView) {

	var newFlt []LoggerContext

	flt := make(map[string]struct{})

	for _, v := range h.filters {
		flt[v.String()] = struct{}{}
	}

	for _, v := range filters {
		if _, ok := flt[v.String()]; !ok {
			newFlt = append(newFlt, v)
		}
	}

	for _, newFilter := range newFlt {
		h.filters = append(h.filters, newFilter)

		nw := tview.NewTextView()
		nw.SetTitle(newFilter.String())

		newWindows = append(newWindows, nw)

		//TODO: this needs to connect to a window
		newHandler := makeNewHandler(newFilter, nw)
		h.filterHandlers = append(h.filterHandlers, newHandler)
	}

	h.multiHandler = log.MultiHandler(h.filterHandlers...)
	return
}
