// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"path/filepath"

	"solod.dev/so/errors"
	"solod.dev/so/fmt"
	"solod.dev/so/os"
	"solod.dev/so/time"
)

func ExampleOpenFile() {
	f, err := os.OpenFile("notes.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}

	fmt.Println("opened", f.Name())
	// Output:
	// opened notes.txt
}

func ExampleOpenFile_append() {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if _, err := f.Write([]byte("appended some data\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}

	fmt.Println("appended to", f.Name())
	// Output:
	// appended to access.log
}

func ExampleChmod() {
	if err := os.Chmod("some-filename", 0644); err != nil {
		panic(err)
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleChtimes() {
	mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
	atime := time.Date(2007, time.March, 2, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes("some-filename", atime, mtime); err != nil {
		panic(err)
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleFileMode() {
	fi, err := os.Lstat("some-filename")
	if err != nil {
		panic(err)
	}

	println("permissions:", fi.Mode().Perm()) // 0o400, 0o777, etc.
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		println("regular file")
	case mode.IsDir():
		println("directory")
	case mode&os.ModeSymlink != 0:
		println("symbolic link")
	case mode&os.ModeNamedPipe != 0:
		println("named pipe")
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleErrNotExist() {
	filename := "a-nonexistent-file"
	if _, err := os.Stat(filename); err == os.ErrNotExist {
		fmt.Println("file does not exist")
	}
}

func ExampleLookupEnv() {
	os.Setenv("SOME_KEY", "value")
	os.Setenv("EMPTY_KEY", "")

	val, ok := os.LookupEnv("SOME_KEY")
	println(val, ok) // value true

	val, ok = os.LookupEnv("EMPTY_KEY")
	println(val, ok) // true

	val, ok = os.LookupEnv("MISSING_KEY")
	println(val, ok) // false

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleGetenv() {
	os.Setenv("NAME", "gopher")
	os.Setenv("BURROW", "/usr/gopher")

	name := os.Getenv("NAME")
	println(name) // gopher

	burrow := os.Getenv("BURROW")
	println(burrow) // /usr/gopher

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleUnsetenv() {
	os.Setenv("TMPDIR", "/my/tmp")
	defer os.Unsetenv("TMPDIR")

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleMkdirTemp() {
	buf := make([]byte, 256)
	dir, err := os.MkdirTemp(buf, "", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(dir) // clean up

	file := filepath.Join(dir, "tmpfile")
	if err := os.WriteFile(file, []byte("content"), 0666); err != nil {
		panic(err)
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleCreateTemp() {
	buf := make([]byte, 256)
	f, err := os.CreateTemp(buf, "", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name()) // clean up

	if _, err := f.Write([]byte("content")); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}

	println(f.Name()) // /tmp/example2094541898

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleReadFile() {
	data, err := os.ReadFile(nil, "testdata/hello")
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleWriteFile() {
	err := os.WriteFile("testdata/hello", []byte("Hello, Gophers!"), 0666)
	if err != nil {
		panic(err)
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleMkdir() {
	err := os.Mkdir("testdir", 0750)
	if err != nil && err != os.ErrExist {
		panic(err)
	}
	err = os.WriteFile("testdir/testfile.txt", []byte("Hello, Gophers!"), 0660)
	if err != nil {
		panic(err)
	}

	fmt.Println("ok")
	// Output:
	// ok
}

func ExampleReadDir() {
	files, err := os.ReadDir(nil, ".")
	if err != nil {
		panic(err)
	}
	defer os.FreeDirEntry(nil, files)

	for _, file := range files {
		fmt.Println(file.Name)
	}
}

func ExampleReadlink() {
	// Create a temporary directory.
	buf := make([]byte, 256)
	d, err := os.MkdirTemp(buf, "", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(d)

	// Write a file in the temporary directory.
	targetPath := filepath.Join(d, "hello.txt")
	if err := os.WriteFile(targetPath, []byte("Hello, Gophers!"), 0644); err != nil {
		panic(err)
	}
	defer os.Remove(targetPath)

	// Create a symbolic link to the file.
	linkPath := filepath.Join(d, "hello.link")
	if err := os.Symlink("hello.txt", filepath.Join(d, "hello.link")); err != nil {
		if err == errors.ErrUnsupported {
			// Allow the example to run on platforms that do not support symbolic links.
			return
		}
		panic(err)
	}
	defer os.Remove(linkPath)

	// Readlink returns the relative path as passed to os.Symlink.
	// It writes into a second buffer, because buf holds the directory path.
	dstBuf := make([]byte, 256)
	dst, err := os.Readlink(dstBuf, linkPath)
	if err != nil {
		panic(err)
	}
	println(dst) // hello.txt

	fmt.Println("ok")
	// Output:
	// ok
}
