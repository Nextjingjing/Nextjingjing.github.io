package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncMigratesAndPreservesContent(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"content/post/demo/index.md":            "---\ntitle: Demo\n---\n\nIntroduction\n\n## สารบัญ\n[old](old)\n\n## Notes\nKeep notes\n{{< page-nav next-link=\"old\" >}}",
		"content/blogs/demo/02-second/index.md": "---\ntitle: Second\n---\nSecond body\n{{< page-nav next-link=\"old\" >}}",
		"content/blogs/demo/01-first/index.md":  "---\ntitle: First\n---\nFirst body\n",
	}
	for path, data := range files {
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := syncPost(root, "demo"); err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for path := range files {
		raw, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		data := string(raw)
		before[path] = data
		if strings.Count(data, "{{< page-nav") != 1 {
			t.Fatalf("duplicate navigation: %s", path)
		}
	}
	post := before["content/post/demo/index.md"]
	for _, want := range []string{"Introduction", "## Notes\nKeep notes", "[บทที่ 1: First](/blogs/demo/01-first/)", "[บทที่ 2: Second](/blogs/demo/02-second/)"} {
		if !strings.Contains(post, want) {
			t.Fatalf("missing %q in %s", want, post)
		}
	}
	if strings.Contains(post, "[old]") {
		t.Fatal("old table remains")
	}
	first := before["content/blogs/demo/01-first/index.md"]
	last := before["content/blogs/demo/02-second/index.md"]
	if !strings.Contains(first, `prev-link="/post/demo/" next-link="/blogs/demo/02-second/"`) || !strings.Contains(first, "First body") {
		t.Fatal(first)
	}
	if !strings.Contains(last, `prev-link="/blogs/demo/01-first/" next-link="/post/demo/"`) || !strings.Contains(last, "Second body") {
		t.Fatal(last)
	}
	if err := syncPost(root, "demo"); err != nil {
		t.Fatal(err)
	}
	for path, data := range before {
		raw, _ := os.ReadFile(filepath.Join(root, path))
		if string(raw) != data {
			t.Fatalf("sync not idempotent: %s", path)
		}
	}
}

func TestCreateRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.md")
	if err := create(path, "ชื่อ: \"ทดสอบ\"", "demo", true); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if titleOf(string(raw)) != "ชื่อ: \"ทดสอบ\"" {
		t.Fatal(string(raw))
	}
	if err := create(path, "Replacement", "demo", true); err == nil {
		t.Fatal("overwrote existing page")
	}
}

func TestNormalizeSlug(t *testing.T) {
	cases := map[string]string{
		"what-is-redis":                   "what-is-redis",
		"what is redis ?":                 "what-is-redis",
		"Redis":                           "redis",
		"  WHAT — IS – REDIS ?  ":         "what-is-redis",
		"\x1b[200~what-is-redis\x1b[201~": "what-is-redis",
		"":                                "",
		"???":                             "",
		"../redis":                        "redis",
	}
	for value, want := range cases {
		if got := normalizeSlug(value); got != want {
			t.Errorf("normalizeSlug(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestAutomaticPostSlug(t *testing.T) {
	root := t.TempDir()
	slug, exists, err := postSlug(root, "Redis")
	if err != nil || exists || slug != "redis" {
		t.Fatalf("%q %v %v", slug, exists, err)
	}
	path := filepath.Join(root, "content/post/redis/index.md")
	if err := create(path, "Redis", "redis", true); err != nil {
		t.Fatal(err)
	}
	slug, exists, err = postSlug(root, "Redis")
	if err != nil || !exists || slug != "redis" {
		t.Fatalf("did not reuse: %q %v %v", slug, exists, err)
	}
	slug, exists, err = postSlug(root, "Redis ?")
	if err != nil || exists || slug != "redis-2" {
		t.Fatalf("collision: %q %v %v", slug, exists, err)
	}
	if got := automaticSlug("อะไรคือ Redis ?", "chapter"); got != "redis" {
		t.Fatal(got)
	}
	if got := automaticSlug("บททดสอบ", "chapter"); got != "chapter" {
		t.Fatal(got)
	}
}
