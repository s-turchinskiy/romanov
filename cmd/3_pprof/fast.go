package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mailru/easyjson"
	"github.com/s-turchinskiy/romanov/cmd/3_pprof/data"
)

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	patternAndroid := regexp.MustCompile("Android")
	patternMSIE := regexp.MustCompile("MSIE")

	seenBrowsers := []string{}
	uniqueBrowsers := 0
	var foundUsersSb strings.Builder

	/*count, err := lineCounter(file)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}*/

	// users := make([]data.User, 0, count)
	scanner := bufio.NewScanner(file)

	var i = -1
	for scanner.Scan() {
		i++
		var user data.User
		// fmt.Printf("%v %v\n", err, line)
		err = easyjson.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			panic(err)
		}
		// users = append(users, user)

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if ok := patternAndroid.MatchString(browser); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browser := range user.Browsers {
			if ok := patternMSIE.MatchString(browser); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !isAndroid || !isMSIE {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user.Email, " [at] ")
		_, _ = fmt.Fprintf(&foundUsersSb, "[%d] %s <%s>\n", i, user.Name, email)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	/*for i, user := range users {


	}*/

	_, _ = fmt.Fprintln(out, "found users:\n"+foundUsersSb.String())
	_, _ = fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

// nolint
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
