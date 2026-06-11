package tui

import "charm.land/bubbles/v2/key"

// keyMap は octobell 独自のキーバインド（list 標準の移動・フィルタ・ヘルプに追加される）。
type keyMap struct {
	Open        key.Binding
	OpenOnly    key.Binding
	MarkRead    key.Binding
	MarkAllRead key.Binding
	Refresh     key.Binding
	Quit        key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Open:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "開く+既読")),
		OpenOnly:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "開くのみ")),
		MarkRead:    key.NewBinding(key.WithKeys("r", "."), key.WithHelp("r", "既読")),
		MarkAllRead: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "全既読")),
		Refresh:     key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "更新")),
		Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "終了")),
	}
}

func (k keyMap) bindings() []key.Binding {
	return []key.Binding{k.Open, k.OpenOnly, k.MarkRead, k.MarkAllRead, k.Refresh, k.Quit}
}
