package ahocorasick

import (
	"sync"
	"testing"
)

func mustFind(t *testing.T, m *Matcher, text, want string) {
	t.Helper()
	got, ok := m.Find(text)
	if !ok {
		t.Fatalf("Find(%q) 未命中，期望 %q", text, want)
	}
	if got != want {
		t.Fatalf("Find(%q) = %q，期望 %q", text, got, want)
	}
}

func mustMiss(t *testing.T, m *Matcher, text string) {
	t.Helper()
	if got, ok := m.Find(text); ok {
		t.Fatalf("Find(%q) = %q，期望未命中", text, got)
	}
}

func TestFindNoWords(t *testing.T) {
	for _, words := range [][]string{nil, {}, {""}, {"", ""}} {
		m := New(words)
		mustMiss(t, m, "任意文本")
		mustMiss(t, m, "")
	}
}

func TestFindSingleWord(t *testing.T) {
	m := New([]string{"赌博"})
	mustFind(t, m, "网络赌博平台", "赌博")
	mustMiss(t, m, "网络博彩平台")
	mustMiss(t, m, "赌 博") // 中间有空格不算命中
}

func TestFindEarliestPosition(t *testing.T) {
	m := New([]string{"foo", "bar"})
	mustFind(t, m, "xbarfoo", "bar") // bar 起点 1 早于 foo 起点 4
}

func TestFindLongerWordEarlierStart(t *testing.T) {
	m := New([]string{"bc", "abcd"})
	mustFind(t, m, "abcd", "abcd") // abcd 起点 0 早于 bc 起点 1（长词后文才结束）
}

func TestFindSameStartPrefersLonger(t *testing.T) {
	m := New([]string{"ab", "abx"})
	mustFind(t, m, "abx", "abx") // 同起点 0，abx 更长
}

func TestFindOverlappingSameEnd(t *testing.T) {
	m := New([]string{"ab", "b"})
	mustFind(t, m, "ab", "ab") // 同一终点取较长词（起点更早）
}

func TestFindDedup(t *testing.T) {
	m := New([]string{"赌", "赌", "赌"})
	mustFind(t, m, "赌", "赌")
}

func TestFindCaseInsensitive(t *testing.T) {
	m := New([]string{"fuck"})
	mustFind(t, m, "FUCK", "fuck")
	mustFind(t, m, "FuCk", "fuck")

	m2 := New([]string{"FUCK"})
	mustFind(t, m2, "fuck", "FUCK") // 命中词保留配置中的原始大小写（便于日志）
	mustFind(t, m2, "Fuck", "FUCK")
}

func TestFindSubstring(t *testing.T) {
	m := New([]string{"abc"})
	mustFind(t, m, "xxabcxx", "abc")
	mustFind(t, m, "abxabc", "abc")
	mustMiss(t, m, "ab")
	mustMiss(t, m, "acb")
}

func TestFindPrefixAndSuffix(t *testing.T) {
	m := New([]string{"foo"})
	mustFind(t, m, "foobar", "foo")
	mustFind(t, m, "xfoo", "foo")
}

func TestFindChineseMultiByte(t *testing.T) {
	m := New([]string{"博彩", "违法"})
	// 这个网站涉嫌违法博彩：违法起点字节 18，博彩起点字节 24
	mustFind(t, m, "这个网站涉嫌违法博彩", "违法")
	mustMiss(t, m, "合法经营")
}

func TestFindFailChainHit(t *testing.T) {
	// abc 与 bc 后缀重叠：处理到位置 2 时需沿 fail 链命中 bc
	m := New([]string{"bc", "abc"})
	mustFind(t, m, "abc", "abc") // abc 起点 0 早于 bc 起点 1
}

func TestFindConcurrent(t *testing.T) {
	m := New([]string{"赌博", "fuck", "博彩"})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				// FUCK 起点 0 早于 赌博（字节 11）
				mustFind(t, m, "FUCK 网络赌博", "fuck")
			}
		}()
	}
	wg.Wait()
}
