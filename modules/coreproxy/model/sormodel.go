package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"github.com/therecipe/qt/core"
)

// SortFilterModel represent a sorted filter model to perform sorting and filtering of the history table
type SortFilterModel struct {
	core.QSortFilterProxyModel

	Custom *CustomTableModel

	filter *Filter

	_ func() `constructor:"init"`

	_ func(column string, order core.Qt__SortOrder) `signal:"sortTableView"`
}

func (m *SortFilterModel) init() {
	m.Custom = NewCustomTableModel(nil)

	m.SetSourceModel(m.Custom)
	//m.SetSortRole(Time)
	//m.Sort(0, core.Qt__DescendingOrder)

	m.ConnectFilterAcceptsRow(m.filterAcceptsRow)
	m.ConnectSortTableView(m.sortTableView)
}

// SetFilter sets a filter on the model
func (m *SortFilterModel) SetFilter(f *Filter) {
	m.filter = f
	m.InvalidateFilter()
}

func (m *SortFilterModel) GetRowId(r int) int64{
	return int64(m.Index(r, 0, core.NewQModelIndex()).Data(ID).ToInt(nil))
}

// ResetFilters remove all filters on the model
func (m *SortFilterModel) ResetFilters() {
	m.InvalidateFilter()
}

func (m *SortFilterModel) filterAcceptsRow(sourceRow int, sourceParent *core.QModelIndex) bool {

	if m.filter == nil {
		return true
	}

	index := m.SourceModel().Index(sourceRow, 0, sourceParent)
	idStr := m.SourceModel().Data(index, 0).ToString()
	id, _ := strconv.Atoi(idStr)

	req, editedReq, resp, editedResp := m.Custom.GetReqResp(int64(id))

	if req == nil {
		return true
	}

	scopeCheck := false
	// scope
	if len(m.filter.Scope) > 0 && m.filter.ScopeOnly {
		for _, s := range m.filter.Scope{
			reg := fmt.Sprintf("^%s$", s)
			r, err := regexp.Compile(reg)
			if err != nil {
				// ignore if the scope is not a valid regexp
				continue
			}
			if r.MatchString(req.Host) {
				scopeCheck = true
				break
			}
		}
		if !scopeCheck {
			return false
		}
	}

	// extension
	showOnlyCheck := false
	if m.filter.Show {
		// show only this stuff
		for _, k := range m.filter.ShowExt{
			if req.Extension == k {
				showOnlyCheck = true
				break
			}
		}
		if !showOnlyCheck {
			return false
		}
	}

	if m.filter.Hide {
		// hide this stuff
		for _, k := range m.filter.HideExt{
			if req.Extension == k {
				return false
			}
		}
	}


	// response status
	next := false //IMP: make me pretier
	for _, i := range m.filter.StatusCode {
		if resp != nil && ((resp.StatusCode <= i+99 && resp.StatusCode >= i) || resp.StatusCode > 599) {
			next = true
			break
		}
	}
	if resp != nil && !next {
		return false
	}

	// text search filter
	txt := ""
	if req != nil {
		txt = req.ToString()
	}
	if editedReq != nil {
		txt = txt + editedReq.ToString()
	}
	if resp != nil {
		txt = txt + resp.ToString()
	}
	if editedResp != nil {
		txt = txt + editedResp.ToString()
	}

	if !strings.Contains(txt, m.filter.Search) {
		return false
	}

	return true

}

func (m *SortFilterModel) sortTableView(column string, order core.Qt__SortOrder) {
	for k, v := range m.Custom.RoleNames() {
		if v.ConstData() == column {
			m.SetSortRole(k)
			m.Sort(0, order)
		}
	}
}
