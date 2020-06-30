package repeater

import (
	"strconv"
	"fmt"
	"regexp"
	"github.com/rhaidiz/broxy/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

var tabNum int

// Gui represents the Gui of the repeater module
type Gui struct {
	core.GuiModule
	Sess *core.Session

	repeaterTabs *widgets.QTabWidget
	tabs         map[int]*TabGui
	tabsMapping	 map[int]int
	tabNum       int
	tabRemoved   bool

	//GoClick func(*TabGui)
	GoClick func(int, string, string, chan string)
	ChangeTabName func(int, string)
	NewTabEvent func(string, string)
	RemoveTabEvent func(int)
	Load func()
	GetStuff func(int, int)(string, string, string)
	_       func(i int) `signal:"changedTab"`
}

// Tab represents a tab in the repeater module
type TabGui struct {
	id				int
	goBtn          *widgets.QPushButton
	cancelBtn      *widgets.QPushButton
	historyPrev		 *widgets.QPushButton
	historyNext		 *widgets.QPushButton
	changeTabName	 *widgets.QPushButton
	HostLine       *widgets.QLineEdit
	TabLine	       *widgets.QLineEdit
	RequestEditor  *widgets.QPlainTextEdit
	ResponseEditor *widgets.QPlainTextEdit
	ComboHistory	 *widgets.QComboBox
	history				 []*tabHistory
}

type tabHistory struct {
	id			int
}

// NewGui creates a new Gui for the repeater module
func NewGui(s *core.Session) *Gui {
	tabNum = 1
	return &Gui{Sess: s, tabNum: 1, tabRemoved: false, tabs: make(map[int]*TabGui), tabsMapping: make(map[int]int) }
}

func (g *Gui) GetSettings() interface{} {
	return nil
}

// GetModuleGui returns the Gui for the current module
func (g *Gui) GetModuleGui() interface{}  {

	g.repeaterTabs = widgets.NewQTabWidget(nil)
	g.Load()
	g.repeaterTabs.SetDocumentMode(true)
	g.repeaterTabs.SetTabsClosable(true)
	g.repeaterTabs.ConnectTabCloseRequested(g.handleClose)
	g.repeaterTabs.ConnectCurrentChanged(g.changedTab)
	//g.repeaterTabs.AddTab(g.NewTab(), strconv.Itoa(g.tabNum))
	g.repeaterTabs.AddTab(widgets.NewQWidget(nil, 0), "+")
	// the following line is to remove the closable button from the last tab
	g.repeaterTabs.TabBar().SetTabButton(g.repeaterTabs.Count()-1, widgets.QTabBar__LeftSide, nil) //.Hide()

	return g.repeaterTabs

}

func (g *Gui) handleClose(index int) {
	w := g.repeaterTabs.Widget(index)
	idLabel := widgets.NewQLabelFromPointer(w.FindChild("mylabel", qtcore.Qt__FindChildrenRecursively).Pointer())
	id, _ := strconv.Atoi(idLabel.Text())

	delete(g.tabs, id)
	g.tabRemoved = true
	g.RemoveTabEvent(id)
	g.repeaterTabs.RemoveTab(index)

}

func (g *Gui) changedTab(i int) {
	if i == g.repeaterTabs.Count()-1 && g.tabRemoved && g.repeaterTabs.Count() > 1 {
		g.repeaterTabs.SetCurrentIndex(i - 1)
	} else if i == g.repeaterTabs.Count()-1 {
		// This branch runs only when a new tab is added with the + button
		// or the first time I load the interface
		g.NewTabEvent("","")
	}
	g.tabRemoved = false
}

// AddNewTab adds a new repeater tab
func (g *Gui) AddNewTab(title string, id int, host, request, response string) {
	g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewTab(title, id, host, request, response), title)
	g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
}

// NewTab adds a new tab
func (g *Gui) NewTab(title string, id int, host, request, response string) widgets.QWidget_ITF {
	t := &TabGui{id: id}
	g.tabs[id] = t

	var label = widgets.NewQLabel(nil, 0)
	label.SetText(fmt.Sprintf("%d", id))
	label.SetObjectName("mylabel")
	label.SetVisible(false)

	mainWidget := widgets.NewQWidget(nil, 0)
	vlayout := widgets.NewQVBoxLayout()
	vlayout.SetContentsMargins(11, 11, 11, 11)

	vlayout.AddWidget(label, 0, 0)
	mainWidget.SetLayout(vlayout)

	hlayout_tab_name := widgets.NewQHBoxLayout()

	t.changeTabName = widgets.NewQPushButton2("Change", nil)
	t.changeTabName.ConnectClicked(func(b bool) {
		newTabName := t.TabLine.Text()
		// delete(g.tabs, tabLabel)
		// g.tabs[newTabName] = t
		g.repeaterTabs.SetTabText(g.repeaterTabs.CurrentIndex(), newTabName)
		g.ChangeTabName(id, newTabName)
	} )

	t.TabLine = widgets.NewQLineEdit(nil)

	hlayout_tab_name.AddWidget(t.changeTabName, 0, 0)
	hlayout_tab_name.AddWidget(t.TabLine, 0, 0)

	vlayout.AddLayout(hlayout_tab_name, 0)

	hlayout := widgets.NewQHBoxLayout()

	t.goBtn = widgets.NewQPushButton2("Go", nil)
	t.cancelBtn = widgets.NewQPushButton2("Cancel", nil)
	t.goBtn.ConnectClicked(func(b bool) {
		c := make(chan string)
		request := t.RequestEditor.ToPlainText()
		var re = regexp.MustCompile(`(?mi)[\r\n]+^accept-encoding:.*$`)
		s := re.ReplaceAllString(request, ``)
		t.RequestEditor.SetPlainText(s)
		host := t.HostLine.Text()
		go g.GoClick(id, host, s, c)
		go func(){
			for resp := range c{
				t.ResponseEditor.SetPlainText(resp)
			}
		}()
	})
	hlayout.AddWidget(t.goBtn, 0, 0)
	hlayout.AddWidget(t.cancelBtn, 0, 0)

	t.HostLine = widgets.NewQLineEdit(nil)
	t.HostLine.SetText(host)
	hlayout.AddWidget(t.HostLine, 0, 0)

	t.ComboHistory = widgets.NewQComboBox(nil)
	t.ComboHistory.AddItems([]string{
	})
	t.ComboHistory.ConnectCurrentIndexChanged(func(i int){
		if len(t.history) > 0 {
			id := t.history[i].id
			host, req, resp := g.GetStuff(t.id, id)
			t.HostLine.SetText(host)
			t.RequestEditor.SetPlainText(req)
			t.ResponseEditor.SetPlainText(resp)
		}
	})

	t.historyPrev = widgets.NewQPushButton2("<", nil)
	t.historyNext = widgets.NewQPushButton2(">", nil)

	t.historyPrev.ConnectClicked(func(b bool) {
			newIndex := t.ComboHistory.CurrentIndex() - 1
			if newIndex >= 0 {
				t.ComboHistory.SetCurrentIndex(newIndex)
			}
	})
	t.historyNext.ConnectClicked(func(b bool) {
			newIndex := t.ComboHistory.CurrentIndex() + 1
			total := t.ComboHistory.Count()
			if newIndex < total {
				t.ComboHistory.SetCurrentIndex(newIndex)
			}
	})

	hlayout.AddWidget(t.historyPrev, 0, 0)
	hlayout.AddWidget(t.historyNext, 0, 0)
	hlayout.AddWidget(t.ComboHistory, 0, 0)

	vlayout.AddLayout(hlayout, 0)

	splitter := widgets.NewQSplitter(nil)
	splitter.SetOrientation(qtcore.Qt__Horizontal)

	t.RequestEditor = widgets.NewQPlainTextEdit(nil)
	t.RequestEditor.SetPlainText(request)
	t.ResponseEditor = widgets.NewQPlainTextEdit(nil)
	t.ResponseEditor.SetReadOnly(true)
	t.ResponseEditor.SetPlainText(response)
	splitter.AddWidget(t.RequestEditor)
	splitter.AddWidget(t.ResponseEditor)

	vlayout.AddWidget(splitter, 0, 0)
	return mainWidget
}

func (g *Gui) AddToHistory(tabId int, idTabContent int, label string) {
	t := g.tabs[tabId]
	t.ComboHistory.AddItem(label, qtcore.NewQVariant())
	h := &tabHistory{ id: idTabContent }
	t.history = append(t.history, h)
	newIndex := t.ComboHistory.Count() - 1
	t.ComboHistory.SetCurrentIndex(newIndex)
}

// Title returns the time of this Gui
func (g *Gui) Title() string {
	return "Repeater"
}
