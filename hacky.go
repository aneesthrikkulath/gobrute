package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yeka/zip"
)

var zip_path = ""
var extract_path = ""
var zip_filename = "sample.zip"

func main() {

	args := os.Args
	for _, vv := range args {
		fmt.Println(vv)
	}
	if len(args) <= 1 {
		fmt.Println("usage: hacky [-h] [-s=START] [-l=LENGTH] [-v] [-a=ALPHABET] [-f=FILE]")
		return
	}

	startPtr := flag.Int("s", 1, "start: where to start. a number from 0 to n")
	stopPtr := flag.Int("l", 5, "length: number of length to try upto. a number from 1 to n")
	verbosePtr := flag.Int("v", 0, "verbose: to show password combinations.0 - none. >= 1, after that many times. ")
	alphabetPtr := flag.String("a", "abc", "alpahabet: character combination to try")
	filePtr := flag.String("f", "", "file: filename with extension (required field)")
	flag.Parse()

	displayTry := *verbosePtr > 0
	fileName := *filePtr

	if fileName == "" {
		fmt.Println("-file name is not provided")
		return
	} else if *startPtr > *stopPtr {
		fmt.Println("start number should not be > length")
		return
	}

	if fi, err := os.Stat(fileName); err == nil {
		//dat, err := ioutil.ReadFile(args[1])
		//checkError(err)
		//fmt.Print(string(dat))
		path, err := os.Getwd()
		checkError(err)
		filename := filepath.Join(path, fi.Name())
		zip_filename = fi.Name()
		zip_path = filename
		extract_path = strings.TrimSuffix(filename, filepath.Ext(filename))
		fmt.Println("File exists :" + filename)
		fmt.Println("Extracting to :" + extract_path)
		fpwd := ""
		start := time.Now()
		ind := 0
		ch := make(chan string, 20)
		for pwd := range GenerateCombinations(*alphabetPtr, *stopPtr, *startPtr) {
			go unlock(ch, filename, pwd)
			if len(ch) >= 15 {
				found := false
				for fpwd = range ch {
					if fpwd != "" {
						found = true
						break
					}
					if len(ch) == 0 {
						break
					}
				}
				if found {
					break
				}
			}
			if displayTry {
				if ind == *verbosePtr {
					fmt.Println("Tried :" + pwd)
					ind = 0
				}
				ind++
			}
		}
		for _ = range ch {
			if len(ch) == 0 {
				break
			}
		}
		elapsed := time.Since(start)
		if fpwd == "" {
			fmt.Println("Password cannot be found!")
		} else {
			fmt.Println("Password is", fpwd)
		}
		fmt.Printf("Time took %s\n", elapsed)
	} else if os.IsNotExist(err) {
		fmt.Println("File not exist")
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}

}

func GenerateCombinations(alphabet string, length int, startlength int) <-chan string {
	c := make(chan string)
	// Starting a separate goroutine that will create all the combinations,
	// feeding them to the channel c
	go func(c chan string) {
		defer close(c) // Once the iteration function is finished, we close the channel

		if startlength <= 0 || startlength == length {
			AppendLetter(c, "", alphabet, length)
		} else if startlength == 1 {
			AddLetterWithouStart(c, "", alphabet, length)
		} else {
			AddLetter(c, "", alphabet, length, startlength)
		}
	}(c)

	return c // Return the channel to the calling function
}

// This function create permutation having same length.
func AppendLetter(c chan string, combo string, alphabet string, length int) {
	if length <= 0 {
		c <- combo
		return
	}

	var newCombo string

	for _, ch := range alphabet {
		newCombo = combo + string(ch)
		AppendLetter(c, newCombo, alphabet, length-1)
	}
}

// AddLetter adds a letter to the combination to create a new combination.
// This new combination is passed on to the channel before we call AddLetter once again
// to add yet another letter to the new combination in case length allows it
func AddLetterWithouStart(c chan string, combo string, alphabet string, length int) {
	// Check if we reached the length limit
	// If so, we just return without adding anything
	if length <= 0 {
		return
	}

	var newCombo string
	for _, ch := range alphabet {
		newCombo = combo + string(ch)
		c <- newCombo
		AddLetterWithouStart(c, newCombo, alphabet, length-1)
	}
}

// AddLetter adds a letter to the combination to create a new combination.
// This new combination is passed on to the channel before we call AddLetter once again
// to add yet another letter to the new combination in case length allows it
func AddLetter(c chan string, combo string, alphabet string, length int, startlength int) {
	// Check if we reached the length limit
	// If so, we just return without adding anything
	if length <= 0 {
		return
	}

	var newCombo string
	for _, ch := range alphabet {
		newCombo = combo + string(ch)
		if len(newCombo) >= startlength {
			c <- newCombo
		}
		AddLetter(c, newCombo, alphabet, length-1, startlength)
	}
}

func unlock(c chan string, filename string, password string) {
	if unzip(filename, password) {
		fmt.Println("Password:" + password)
		c <- password
	} else {
		c <- ""
	}
}
func unzip(filename string, password string) bool {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return false
	}
	defer r.Close()

	buffer := new(bytes.Buffer)

	for _, f := range r.File {
		f.SetPassword(password)
		r, err := f.Open()
		if err != nil {
			return false
		}
		defer r.Close()
		n, err := io.Copy(buffer, r)
		if n == 0 || err != nil {
			return false
		}
		break
	}
	return true
}

func checkFor7Zip() bool {
	_, e := exec.LookPath("7z")
	if e != nil {
		return false
	}
	checkError(e)
	return true
}

func extractZipWithPassword(zip_password string) bool {
	//fmt.Printf("Unzipping `%s` to directory `%s`\n", zip_path, extract_path)
	//commandString := fmt.Sprintf(`7za e %s -o%s -p"%s" -aoa`, zip_path, extract_path, zip_password)
	commandString := fmt.Sprintf(`7z e %s -o%s -p"%s" -aoa`, zip_path, extract_path, zip_password)
	commandSlice := strings.Fields(commandString)
	//fmt.Println(commandString)
	c := exec.Command(commandSlice[0], commandSlice[1:]...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		if fmt.Sprint(err) == "exit status 2" {
			return false
		} else {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return false
		}
	}
	return true
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
