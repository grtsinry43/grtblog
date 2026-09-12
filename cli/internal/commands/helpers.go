package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
)

// taxonomyIndex 缓存分类/专栏/标签，用于名称 ↔ ID 解析。
type taxonomyIndex struct {
	Categories []client.Category
	Columns    []client.Column
	Tags       []client.Tag
}

func loadTaxonomy(c *client.Client) (*taxonomyIndex, error) {
	idx := &taxonomyIndex{}
	if err := c.Get("/categories", nil, &idx.Categories); err != nil {
		return nil, err
	}
	if err := c.Get("/columns", nil, &idx.Columns); err != nil {
		return nil, err
	}
	if err := c.Get("/tags", nil, &idx.Tags); err != nil {
		return nil, err
	}
	return idx, nil
}

// resolveID 将「名称或数字 ID」解析为数字 ID。
func resolveID(nameOrID string, names map[string]int64, kind string) (int64, error) {
	s := strings.TrimSpace(nameOrID)
	if s == "" {
		return 0, fmt.Errorf("%s不能为空", kind)
	}
	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		return id, nil
	}
	if id, ok := names[strings.ToLower(s)]; ok {
		return id, nil
	}
	available := make([]string, 0, len(names))
	for n := range names {
		available = append(available, n)
	}
	sort.Strings(available)
	return 0, fmt.Errorf("未找到%s %q（可用: %s）", kind, s, strings.Join(available, ", "))
}

func (t *taxonomyIndex) categoryNames() map[string]int64 {
	m := make(map[string]int64, len(t.Categories))
	for _, c := range t.Categories {
		m[strings.ToLower(c.Name)] = c.ID
	}
	return m
}

func (t *taxonomyIndex) columnNames() map[string]int64 {
	m := make(map[string]int64, len(t.Columns))
	for _, c := range t.Columns {
		m[strings.ToLower(c.Name)] = c.ID
	}
	return m
}

func (t *taxonomyIndex) tagNames() map[string]int64 {
	m := make(map[string]int64, len(t.Tags))
	for _, tag := range t.Tags {
		m[strings.ToLower(tag.Name)] = tag.ID
	}
	return m
}

// parseIDs 将命令行参数解析为 int64 ID 列表。
func parseIDs(args []string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, a := range args {
		id, err := strconv.ParseInt(strings.TrimSpace(a), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("无效的 ID %q", a)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// resolveIDs 批量解析名称/ID 混合列表。
func resolveIDs(list []string, names map[string]int64, kind string) ([]int64, error) {
	if len(list) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(list))
	for _, item := range list {
		id, err := resolveID(item, names, kind)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
