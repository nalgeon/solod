package main

import (
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
}
