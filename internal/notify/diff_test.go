package notify

import (
	"reflect"
	"testing"
)

func TestDiffer(t *testing.T) {
	d := NewDiffer()

	// 初回はバックログ全件を通知しない（新着なし）。
	if got := d.New([]string{"a", "b"}); len(got) != 0 {
		t.Fatalf("初回は新着なしを期待, got %v", got)
	}
	// 既知のみ → 新着なし。
	if got := d.New([]string{"a", "b"}); len(got) != 0 {
		t.Fatalf("既知のみは新着なしを期待, got %v", got)
	}
	// 新規 c のみ新着。
	if got := d.New([]string{"a", "b", "c"}); !reflect.DeepEqual(got, []string{"c"}) {
		t.Fatalf("新着 [c] を期待, got %v", got)
	}
	// 再度 c は既知。
	if got := d.New([]string{"c"}); len(got) != 0 {
		t.Fatalf("c は既知のはず, got %v", got)
	}
}
