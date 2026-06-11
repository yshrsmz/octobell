package notify

// Differ は前回ポーリングで見たスレッド ID を記憶し、新規分だけを返す。
// OS 通知を「新着のみ」に絞り、通知のスパム化を防ぐために使う。
type Differ struct {
	seen map[string]struct{}
}

// NewDiffer は空の Differ を生成する。
func NewDiffer() *Differ {
	return &Differ{seen: make(map[string]struct{})}
}

// New は ids のうち未知（新着）のものを返し、全 ids を記憶する。
// 初回呼び出し（記憶が空）のときは、既存バックログ全件を通知しないよう新着なしとして扱う。
func (d *Differ) New(ids []string) []string {
	first := len(d.seen) == 0
	var fresh []string
	for _, id := range ids {
		if _, ok := d.seen[id]; !ok && !first {
			fresh = append(fresh, id)
		}
	}
	for _, id := range ids {
		d.seen[id] = struct{}{}
	}
	return fresh
}
