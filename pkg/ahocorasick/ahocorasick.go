// Package ahocorasick 提供 Aho-Corasick 多模式串匹配算法实现。
// 构建一次后自动机只读，可被多个 goroutine 并发调用 Find。
package ahocorasick

import "strings"

// node 是自动机的 trie 节点。
type node struct {
	children map[byte]*node // 子节点（按字节索引，UTF-8 安全）
	fail     *node          // 失配指针：当前路径的最长真后缀节点
	word     string         // 非空表示该节点是某个词条的终点（保留配置中的原始大小写）
}

// Matcher 是编译完成的 Aho-Corasick 自动机，构建后只读、可并发使用。
type Matcher struct {
	root *node
}

// New 根据词条列表构建自动机。
// 空串与重复词条会被忽略；构建与匹配均按小写化字节进行（大小写不敏感），
// 但 Find 返回的命中词保留配置中的原始大小写，便于日志展示。
// 构建复杂度 O(总词条长度)，单次匹配复杂度 O(文本长度)。
func New(words []string) *Matcher {
	root := &node{children: make(map[byte]*node)}

	// 第一阶段：构建 trie（按小写化词条建边），终点记录原始大小写词条
	for _, w := range words {
		if w == "" {
			continue
		}
		lowered := strings.ToLower(w) // 与 Find 的文本小写化保持一致
		cur := root
		for i := 0; i < len(lowered); i++ {
			c := lowered[i]
			next, ok := cur.children[c]
			if !ok {
				next = &node{children: make(map[byte]*node)}
				cur.children[c] = next
			}
			cur = next
		}
		if cur.word == "" {
			cur.word = w // 重复词条只保留一次（以先出现者的大小写为准）
		}
	}

	// 第二阶段：BFS 计算 fail 指针（根的直接子节点 fail 指向根）
	queue := make([]*node, 0, 16)
	for _, child := range root.children {
		child.fail = root
		queue = append(queue, child)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for c, child := range cur.children {
			f := cur.fail
			for f != nil {
				if next, ok := f.children[c]; ok {
					child.fail = next
					break
				}
				f = f.fail
			}
			if child.fail == nil {
				child.fail = root
			}
			queue = append(queue, child)
		}
	}

	return &Matcher{root: root}
}

// Find 返回文本中起点最早（出现位置最靠前）的命中词条；
// 起点相同时返回较长词条。匹配前文本统一小写化。
// 未命中返回 ("", false)。
func (m *Matcher) Find(text string) (string, bool) {
	if m == nil || m.root == nil || len(m.root.children) == 0 {
		return "", false
	}

	text = strings.ToLower(text)
	cur := m.root
	bestWord := ""
	bestStart := -1
	for i := 0; i < len(text); i++ {
		c := text[i]
		// 沿 fail 链回退，直到找到可转移的节点（或回到根）
		for cur != m.root {
			if _, ok := cur.children[c]; ok {
				break
			}
			cur = cur.fail
		}
		if next, ok := cur.children[c]; ok {
			cur = next
		}
		// 检查当前节点及 fail 链上的词条终点。
		// 同一终点沿链向下词长递减、起点递增，首个命中即该位置起点最早者；
		// 更长词条可能在后文结束（起点更早），故必须扫描完整段文本取全局最小起点。
		for n := cur; n != nil; n = n.fail {
			if n.word != "" {
				start := i - len(n.word) + 1
				// <=：同起点（更长词在后文才结束）时替换为较长的命中词；
				// 同一终点沿 fail 链向下只会更短、起点更晚，不会误选。
				if bestStart < 0 || start <= bestStart {
					bestStart = start
					bestWord = n.word
				}
				break
			}
		}
	}
	return bestWord, bestStart >= 0
}
