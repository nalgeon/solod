package main

import (
	"github.com/nalgeon/solod/so/io"
	"github.com/nalgeon/solod/so/mem"
	"github.com/nalgeon/solod/so/os"
)

func main() {
	{
		// WriteFile, ReadFile.
		name := "test_rw.txt"
		data := []byte("hello world")
		err := os.WriteFile(name, data)
		if err != nil {
			panic("WriteFile failed")
		}
		defer os.Remove(name)

		b, err := os.ReadFile(nil, name)
		if err != nil {
			panic("ReadFile failed")
		}
		defer mem.FreeSlice(nil, b)
		if string(b) != string(data) {
			panic("ReadFile: wrong data")
		}
	}
	{
		// Create, Write, Close.
		name := "test_file.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		defer os.Remove(name)

		// Write.
		n, err := f.Write([]byte("abcdef"))
		if err != nil {
			panic("Write failed")
		}
		if n != 6 {
			panic("Write: wrong count")
		}

		// Close.
		err = f.Close()
		if err != nil {
			panic("Close failed")
		}
	}
	{
		// Open, Read, Close.
		name := "test_file.txt"
		data := []byte("abcdef")
		err := os.WriteFile(name, data)
		if err != nil {
			panic("WriteFile failed")
		}
		defer os.Remove(name)

		// Open.
		f, err := os.Open(name)
		if err != nil {
			panic("Open failed")
		}

		// Read.
		buf := make([]byte, 10)
		n, err := f.Read(buf)
		if err != nil {
			panic("Read failed")
		}
		if n != 6 {
			panic("Read: wrong count")
		}
		if string(buf[:n]) != "abcdef" {
			panic("Read: wrong data")
		}

		// Close.
		err = f.Close()
		if err != nil {
			panic("Close failed")
		}
	}
	{
		// Remove.
		name := "test_remove.txt"
		err := os.WriteFile(name, []byte("tmp"))
		if err != nil {
			panic("WriteFile failed")
		}
		err = os.Remove(name)
		if err != nil {
			panic("Remove failed")
		}
		_, err = os.Open(name)
		if err == nil {
			panic("Open after Remove should fail")
		}
	}
	{
		// Rename.
		oldName := "test_old.txt"
		newName := "test_new.txt"
		os.WriteFile(oldName, []byte("renamed"))
		err := os.Rename(oldName, newName)
		if err != nil {
			panic("Rename failed")
		}
		defer os.Remove(newName)
		b, err := os.ReadFile(nil, newName)
		if err != nil {
			panic("ReadFile after Rename failed")
		}
		defer mem.FreeSlice(nil, b)
		if string(b) != "renamed" {
			panic("Rename: wrong data")
		}
	}
	{
		// ErrNotExist.
		_, err := os.Open("nonexistent_file.txt")
		if err != os.ErrNotExist {
			panic("Open nonexistent: wrong error")
		}
	}
	{
		// Seek.
		name := "test_seek.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		defer os.Remove(name)
		f.Write([]byte("abcdef"))
		pos, err := f.Seek(0, io.SeekStart)
		if err != nil {
			panic("Seek failed")
		}
		if pos != 0 {
			panic("Seek: wrong position")
		}
		buf := make([]byte, 6)
		n, err := f.Read(buf)
		if err != nil {
			panic("Read after Seek failed")
		}
		if string(buf[:n]) != "abcdef" {
			panic("Seek: wrong data")
		}
		f.Close()
	}
	{
		// ReadAt.
		name := "test_readat.txt"
		err := os.WriteFile(name, []byte("hello world"))
		if err != nil {
			panic("WriteFile failed")
		}
		defer os.Remove(name)
		f, err := os.Open(name)
		if err != nil {
			panic("Open failed")
		}
		buf := make([]byte, 5)
		n, err := f.ReadAt(buf, 6)
		if err != nil {
			panic("ReadAt failed")
		}
		if n != 5 {
			panic("ReadAt: wrong count")
		}
		if string(buf[:n]) != "world" {
			panic("ReadAt: wrong data")
		}
		f.Close()
	}
	{
		// WriteAt.
		name := "test_writeat.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		defer os.Remove(name)
		f.Write([]byte("hello world"))
		_, err = f.WriteAt([]byte("WORLD"), 6)
		if err != nil {
			panic("WriteAt failed")
		}
		f.Close()

		b, err := os.ReadFile(nil, name)
		if err != nil {
			panic("ReadFile failed")
		}
		defer mem.FreeSlice(nil, b)
		if string(b) != "hello WORLD" {
			panic("WriteAt: wrong data")
		}
	}
	{
		// WriteString.
		name := "test_writestr.txt"
		f, err := os.Create(name)
		if err != nil {
			panic("Create failed")
		}
		defer os.Remove(name)
		n, err := f.WriteString("hello")
		if err != nil {
			panic("WriteString failed")
		}
		if n != 5 {
			panic("WriteString: wrong count")
		}
		f.Close()

		b, err := os.ReadFile(nil, name)
		if err != nil {
			panic("ReadFile failed")
		}
		defer mem.FreeSlice(nil, b)
		if string(b) != "hello" {
			panic("WriteString: wrong data")
		}
	}
	{
		// Getenv.
		path := os.Getenv("PATH")
		if len(path) == 0 {
			panic("Getenv PATH: empty")
		}
	}
}
