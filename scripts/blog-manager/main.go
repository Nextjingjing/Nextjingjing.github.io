package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var input = bufio.NewReader(os.Stdin)
var navRE = regexp.MustCompile(`(?s)\{\{<\s*page-nav\b.*?>\}\}`)
var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const start = "<!-- blog-manager:toc:start -->"
const end = "<!-- blog-manager:toc:end -->"

type page struct{ dir, title string }

func ask(prompt string) string {
	fmt.Print(prompt + ": ")
	line, err := input.ReadString('\n')
	if err != nil && len(line) == 0 {
		fmt.Println()
		os.Exit(0)
	}
	return strings.TrimSpace(strings.NewReplacer("\x1b[200~", "", "\x1b[201~", "").Replace(line))
}

type option struct{ label, value string }

// Read and restore the actual terminal flags, without relying on stty variants.
func terminalState(state *syscall.Termios, request uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), request, uintptr(unsafe.Pointer(state)))
	if errno != 0 {
		return errno
	}
	return nil
}

func choose(title string, options []option) (string, error) {
	var state syscall.Termios
	err := terminalState(&state, syscall.TCGETS)
	if err != nil {
		fmt.Println("\n" + title)
		for i, item := range options {
			fmt.Printf("%d. %s\n", i+1, item.label)
		}
		for {
			value := ask("เลือกหมายเลข หรือคำสั่ง")
			n, err := strconv.Atoi(value)
			if err == nil && n >= 1 && n <= len(options) {
				return options[n-1].value, nil
			}
			for _, item := range options {
				if value == item.value {
					return value, nil
				}
			}
		}
	}
	raw := state
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := terminalState(&raw, syscall.TCSETS); err != nil {
		return "", err
	}
	defer terminalState(&state, syscall.TCSETS)
	// Keep menu redraws separate from prompts, typed text, and status messages.
	fmt.Print("\033[?1049h\033[?25l")
	defer fmt.Print("\033[?25h\033[?1049l")
	selected := 0
	draw := func() {
		fmt.Print("\033[H\033[2J")
		fmt.Println(title + " — ↑ ↓ เลือก • Enter ยืนยัน • Ctrl+C ออก")
		for i, item := range options {
			if i == selected {
				fmt.Printf("\033[36m❯ %s\033[0m\n", item.label)
			} else {
				fmt.Printf("  %s\n", item.label)
			}
		}
	}
	draw()
	for {
		key, err := input.ReadByte()
		if err != nil {
			return "q", nil
		}
		switch key {
		case 3, 4:
			return "q", nil
		case '\r', '\n':
			fmt.Println()
			return options[selected].value, nil
		case 27:
			// CSI and application cursor sequences both end in A/B.
			prefix, err := input.ReadByte()
			if err != nil {
				return "q", nil
			}
			if prefix != '[' && prefix != 'O' {
				continue
			}
			direction, err := input.ReadByte()
			if err != nil {
				return "q", nil
			}
			if direction == 'A' {
				selected = (selected + len(options) - 1) % len(options)
			}
			if direction == 'B' {
				selected = (selected + 1) % len(options)
			}
			draw()
		default:
			for _, item := range options {
				if item.value == string(key) {
					return item.value, nil
				}
			}
		}
	}
}

func titleOf(data string) string {
	for _, line := range strings.Split(data, "\n") {
		if strings.HasPrefix(line, "title:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			if decoded, err := strconv.Unquote(value); err == nil {
				return decoded
			}
			return strings.Trim(value, "'\"")
		}
	}
	return ""
}
func pages(root string) ([]page, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []page
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name(), "index.md"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, page{entry.Name(), titleOf(string(data))})
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := strconv.Atoi(strings.SplitN(result[i].dir, "-", 2)[0])
		b, _ := strconv.Atoi(strings.SplitN(result[j].dir, "-", 2)[0])
		if a != b {
			return a < b
		}
		return result[i].dir < result[j].dir
	})
	return result, nil
}
func write(path, data string) error {
	// Rename a temporary file so an interrupted write cannot truncate a page.
	f, err := os.CreateTemp(filepath.Dir(path), ".blog-manager-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.WriteString(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Chmod(0644); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func navigation(data, prev, next string) string {
	data = strings.TrimSpace(navRE.ReplaceAllString(data, ""))
	return data + fmt.Sprintf("\n\n{{< page-nav prev-link=%q next-link=%q >}}\n", prev, next)
}
func syncPost(root, slug string) error {
	chapters, err := pages(filepath.Join(root, "content/blogs", slug))
	if err != nil {
		return err
	}
	path := filepath.Join(root, "content/post", slug, "index.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data := string(raw)
	var toc strings.Builder
	toc.WriteString(start + "\n## สารบัญ\n\n")
	for i, ch := range chapters {
		label := strings.NewReplacer("[", "\\[", "]", "\\]", "\n", " ").Replace(ch.title)
		fmt.Fprintf(&toc, "- [บทที่ %d: %s](/blogs/%s/%s/)\n", i+1, label, slug, ch.dir)
	}
	toc.WriteString(end)
	a, b := strings.Index(data, start), strings.Index(data, end)
	if a >= 0 && b >= a {
		data = data[:a] + toc.String() + data[b+len(end):]
	} else if a >= 0 || b >= 0 {
		return fmt.Errorf("เครื่องหมายสารบัญไม่ครบใน %s", path)
	} else {
		// Migrate the existing heading, preserving subsequent sections and prose.
		heading := regexp.MustCompile(`(?m)^## สารบัญ\s*$`).FindStringIndex(data)
		if heading != nil {
			tail := data[heading[1]:]
			cut := len(tail)
			if next := regexp.MustCompile(`(?m)^#{1,2} `).FindStringIndex(tail); next != nil {
				cut = next[0]
			}
			if next := navRE.FindStringIndex(tail); next != nil && next[0] < cut {
				cut = next[0]
			}
			data = data[:heading[0]] + toc.String() + "\n\n" + tail[cut:]
		} else {
			data = strings.TrimSpace(data) + "\n\n" + toc.String()
		}
	}
	postURL := "/post/" + slug + "/"
	next := postURL
	if len(chapters) > 0 {
		next = "/blogs/" + slug + "/" + chapters[0].dir + "/"
	}
	if err := write(path, navigation(data, "/", next)); err != nil {
		return err
	}
	for i, ch := range chapters {
		path := filepath.Join(root, "content/blogs", slug, ch.dir, "index.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		prev, next := postURL, postURL
		if i > 0 {
			prev = "/blogs/" + slug + "/" + chapters[i-1].dir + "/"
		}
		if i+1 < len(chapters) {
			next = "/blogs/" + slug + "/" + chapters[i+1].dir + "/"
		}
		if err := write(path, navigation(string(raw), prev, next)); err != nil {
			return err
		}
	}
	return nil
}
func create(path, title, slug string, post bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	category := "Blogs"
	extra := ""
	if post {
		category = "Programming"
		extra = fmt.Sprintf("slug: %q\n", slug)
	}
	_, err = fmt.Fprintf(f, "---\ntitle: %q\ndescription: %q\ndate: %s\n%scategories:\n    - %s\ntoc: true\n---\n\n## %s\n\nเขียนเนื้อหาที่นี่\n", title, title, time.Now().Format("2006-01-02"), extra, category, title)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func normalizeSlug(value string) string {
	// Pasted input can carry terminal bracketed-paste delimiters.
	value = strings.NewReplacer("\x1b[200~", "", "\x1b[201~", "").Replace(value)
	value = strings.ToLower(value)
	var result strings.Builder
	separator := false
	for _, ch := range value {
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(ch)
			separator = false
		} else {
			separator = true
		}
	}
	return result.String()
}

func automaticSlug(title, fallback string) string {
	if slug := normalizeSlug(title); slug != "" {
		return slug
	}
	return fallback
}

func postSlug(root, title string) (string, bool, error) {
	base := automaticSlug(title, "post")
	for i := 1; ; i++ {
		slug := base
		if i > 1 {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		path := filepath.Join(root, "content/post", slug)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return slug, false, nil
		} else if err != nil {
			return "", false, err
		}
		raw, err := os.ReadFile(filepath.Join(path, "index.md"))
		if err == nil && strings.EqualFold(titleOf(string(raw)), title) {
			return slug, true, nil
		}
	}
}

func edit(path string) error {
	fmt.Println("เปิดไฟล์:", path)
	editor := os.Getenv("EDITOR")
	var cmd *exec.Cmd
	if editor != "" {
		cmd = exec.Command("sh", "-c", editor+` "$1"`, "blog-manager", path)
	} else if code, err := exec.LookPath("code"); err == nil {
		cmd = exec.Command(code, "--reuse-window", "--wait", path)
	} else {
		fmt.Println("เปิดไฟล์ตาม path ด้านบนเพื่อเขียนเนื้อหา แล้วบันทึก")
		ask("กด Enter เมื่อเขียนเสร็จ เพื่ออัปเดตสารบัญ")
		return nil
	}
	fmt.Println("เขียนเนื้อหาใน editor → บันทึก → ปิดแท็บไฟล์นี้ เพื่อกลับเมนูและอัปเดตสารบัญ")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("เปิด editor ไม่สำเร็จ ไฟล์อยู่ที่:", path)
		ask("กด Enter เพื่อกลับเมนู")
		return err
	}
	return nil
}
func run(root string) error {
	// Read plain pasted text rather than terminal paste-control sequences.
	fmt.Print("\033[?2004l")
	for {
		posts, err := pages(filepath.Join(root, "content/post"))
		if err != nil {
			return err
		}
		options := make([]option, 0, len(posts)+2)
		for i, p := range posts {
			options = append(options, option{p.title + " (" + p.dir + ")", strconv.Itoa(i + 1)})
		}
		options = append(options, option{"＋ สร้างโพสต์ใหม่", "n"}, option{"ออก", "q"})
		choice, err := choose("1. เลือกโพสต์ที่จะเขียน", options)
		if err != nil {
			return err
		}
		if choice == "q" {
			return nil
		}
		slug := ""
		if choice == "n" {
			title := ask("ชื่อโพสต์ (Enter ว่าง = กลับ)")
			if title == "" {
				continue
			}
			var existing bool
			slug, existing, err = postSlug(root, title)
			if err != nil {
				return err
			}
			if !existing {
				if err := create(filepath.Join(root, "content/post", slug, "index.md"), title, slug, true); err != nil {
					return err
				}
			}

		} else {
			n, err := strconv.Atoi(choice)
			if err != nil || n < 1 || n > len(posts) {
				continue
			}
			slug = posts[n-1].dir
		}
		for {
			if err := syncPost(root, slug); err != nil {
				return err
			}
			chapters, err := pages(filepath.Join(root, "content/blogs", slug))
			if err != nil {
				return err
			}
			options := make([]option, 0, len(chapters)+4)
			for i, ch := range chapters {
				options = append(options, option{fmt.Sprintf("บทที่ %d — %s (เปิดเขียน)", i+1, ch.title), strconv.Itoa(i + 1)})
			}
			options = append(options, option{"＋ สร้างบทความใหม่", "n"}, option{"แก้บทนำของโพสต์ / ดูสารบัญ", "p"}, option{"รีเฟรชสารบัญจากไฟล์ที่บันทึก", "s"}, option{"กลับไปเลือกโพสต์", "b"})
			choice, err := choose(fmt.Sprintf("2. %s — %d บท (สารบัญและปุ่มถัดไปอัปเดตแล้ว)", slug, len(chapters)), options)
			if err != nil {
				return err
			}
			if choice == "q" {
				return nil
			}
			if choice == "b" {
				break
			}
			switch choice {
			case "s":
				fmt.Println("อัปเดตสารบัญและปุ่มนำทางแล้ว")
			case "p":
				if err := edit(filepath.Join(root, "content/post", slug, "index.md")); err != nil {
					fmt.Println(err)
				}
			case "n":
				title := ask("ชื่อบทความ (Enter ว่าง = กลับ)")
				if title == "" {
					continue
				}
				chSlug := automaticSlug(title, "chapter")
				number := 0
				for _, ch := range chapters {
					n, _ := strconv.Atoi(strings.SplitN(ch.dir, "-", 2)[0])
					if n > number {
						number = n
					}
				}
				path := filepath.Join(root, "content/blogs", slug, fmt.Sprintf("%02d-%s", number+1, chSlug), "index.md")
				if err := create(path, title, slug, false); err != nil {
					fmt.Println(err)
					continue
				}
				if err := syncPost(root, slug); err != nil {
					return err
				}
				if err := edit(path); err != nil {
					fmt.Println(err)
				}
			default:
				n, err := strconv.Atoi(choice)
				if err == nil && n >= 1 && n <= len(chapters) {
					if err := edit(filepath.Join(root, "content/blogs", slug, chapters[n-1].dir, "index.md")); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if info, err := os.Stat(filepath.Join(root, "content")); err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			fmt.Fprintln(os.Stderr, "กรุณารันในโปรเจกต์ Hugo")
			os.Exit(1)
		}
		root = parent
	}
	if len(os.Args) == 3 && os.Args[1] == "sync" {
		if !slugRE.MatchString(os.Args[2]) {
			fmt.Fprintln(os.Stderr, "Slug ไม่ถูกต้อง")
			os.Exit(1)
		}
		err = syncPost(root, os.Args[2])
	} else if len(os.Args) == 1 {
		err = run(root)
	} else {
		err = fmt.Errorf("ใช้ blog-manager หรือ blog-manager sync <slug>")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
