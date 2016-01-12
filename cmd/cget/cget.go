// Steve Phillips / elimisteve
// 2015.12.23

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/types"
)

var (
	db backend.Backend
)

func init() {
	fs, err := backend.LoadOrCreateFileSystem(
		os.Getenv("CRYPTAG_BACKEND_PATH"),
		os.Getenv("CRYPTAG_BACKEND_NAME"),
	)
	if err != nil {
		log.Fatalf("LoadFileSystem error: %v\n", err)
	}

	db = fs
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	switch os.Args[1] {
	// TODO: Add "decrypt" case
	// TODO: Add "file" case

	case "list":
		plaintags := append(os.Args[2:], "type:file")

		rows, err := db.ListRows(plaintags)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range rows {
			fmt.Printf("%v\t\t%v\n\n", types.RowTagWithPrefix(r, "filename:"),
				strings.Join(r.PlainTags(), "  "))
		}

	case "tags":
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}
		for _, pair := range pairs {
			fmt.Printf("%s  %s\n", pair.Random, pair.Plain())
		}

	default: // Decrypt, save to ~/.cryptag/decrypted/(filename from filename:...)
		plaintags := os.Args[1:]

		// TODO: Temporary?
		plaintags = append(plaintags, "type:file")

		rows, err := db.RowsFromPlainTags(plaintags)
		if err != nil {
			log.Fatal(err)
		}

		if len(rows) == 0 {
			log.Fatal(types.ErrRowsNotFound)
		}

		var rowFilename string
		for _, r := range rows {
			if rowFilename, err = saveRowAsFile(r); err != nil {
				log.Printf("Error locally saving file: %v\n", err)
				continue
			}
			log.Printf("Saved new file: %v\n", rowFilename)
		}
	}
}

func saveRowAsFile(r *types.Row) (filepath string, err error) {
	filename := types.RowTagWithPrefix(r, "filename:")
	if filename == "" {
		filename = fmt.Sprintf("%v", time.Now().Unix())
	}

	dir := path.Join(cryptag.TrustedBasePath, "decrypted")
	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Error creating directory `%v`: %v", dir, err)
	}

	filepath = path.Join(dir, filename)
	return filepath, ioutil.WriteFile(filepath, r.Decrypted(), 0644)
}

var (
	usage = "Usage: " + filepath.Base(os.Args[0]) + " tag1 [tag2 ...]"
)