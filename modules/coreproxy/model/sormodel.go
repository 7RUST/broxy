package model

import (
	"github.com/therecipe/qt/core"
)

type SortFilterModel struct {
	core.QSortFilterProxyModel

	Custom *CustomTableModel

	_ func() `constructor:"init"`

	_ func(column string, order core.Qt__SortOrder) `signal:"sortTableView"`
}

func init() {
	CustomTableModel_QmlRegisterType2("CustomQmlTypes", 1, 0, "SortFilterModel")
}

func (m *SortFilterModel) init() {
	m.Custom = NewCustomTableModel(nil)

	m.SetSourceModel(m.Custom)
	//m.SetSortRole(Time)
	//m.Sort(0, core.Qt__DescendingOrder)

	m.ConnectSortTableView(m.sortTableView)
}

func (m *SortFilterModel) sortTableView(column string, order core.Qt__SortOrder) {
	for k, v := range m.Custom.RoleNames() {
		if v.ConstData() == column {
			m.SetSortRole(k)
			m.Sort(0, order)
		}
	}
}
