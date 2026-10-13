"""Linux PTY regression: menu -> visible text input -> menu -> restored terminal."""
import fcntl
import os
from pathlib import Path
import pty
import select
import subprocess
import tempfile
import termios
import time

source = Path(__file__).with_name("main.go")
with tempfile.TemporaryDirectory() as directory:
    root = Path(directory)
    (root / "content").mkdir()
    executable = root / "blog-manager"
    subprocess.run(["go", "build", "-o", str(executable), str(source)], check=True)
    fake_bin = root / "bin"
    fake_bin.mkdir()
    fake_code = fake_bin / "code"
    fake_code.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > "' + str(root / "editor-args") + '"\n')
    fake_code.chmod(0o755)
    master, slave = pty.openpty()
    original = termios.tcgetattr(slave)

    def setup():
        os.setsid()
        fcntl.ioctl(0, termios.TIOCSCTTY, 0)

    process = subprocess.Popen(
        [str(executable)], cwd=root, stdin=slave, stdout=slave, stderr=slave,
        preexec_fn=setup, env={**os.environ, "EDITOR": "", "PATH": str(fake_bin)+os.pathsep+os.environ["PATH"]},
    )

    def expect(text):
        output = b""
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline:
            if select.select([master], [], [], 0.1)[0]:
                output += os.read(master, 65536)
                if text.encode() in output:
                    return output
        raise AssertionError(f"Missing {text!r}: {output!r}")

    def type_visible(text):
        flags = termios.tcgetattr(slave)[3]
        assert flags & termios.ECHO and flags & termios.ICANON
        os.write(master, text.encode())
        expect(text)
        os.write(master, b"\r")

    try:
        expect("❯ ＋ สร้างโพสต์ใหม่")
        os.write(master, b"\r")
        expect("ชื่อโพสต์ (Enter ว่าง = กลับ): ")
        assert termios.tcgetattr(slave) == original
        type_visible("โพสต์ทดสอบ")
        expect("❯ ＋ สร้างบทความใหม่")
        os.write(master, b"\r")
        expect("ชื่อบทความ (Enter ว่าง = กลับ): ")
        type_visible("บททดสอบ")
        expect("❯ บทที่ 1 — บททดสอบ (เปิดเขียน)")
        os.write(master, b"\x1b[A\r")  # wrap to Back
        expect("❯ โพสต์ทดสอบ")
        os.write(master, b"\x1b[A\r")  # wrap to Exit
        process.wait(timeout=5)
        assert process.returncode == 0
        assert termios.tcgetattr(slave) == original
        assert 'title: "โพสต์ทดสอบ"' in (root / "content/post/post/index.md").read_text()
        assert 'title: "บททดสอบ"' in (root / "content/blogs/post/01-chapter/index.md").read_text()
        args = (root / "editor-args").read_text().splitlines()
        assert args == ["--reuse-window", "--wait", str(root / "content/blogs/post/01-chapter/index.md")], args
        post = (root / "content/post/post/index.md").read_text()
        assert "/blogs/post/01-chapter/" in post
        print("PASS: arrow menus, visible Thai typing, post/chapter creation, terminal restored")
    finally:
        if process.poll() is None:
            process.terminate()
            process.wait(timeout=5)
        os.close(master)
        os.close(slave)
