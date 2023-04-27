/* This edits strings in comments and descriptions to add links */
package main

import (
	"fmt"
	"html"
	"regexp"
)

type fixstring struct {
	/* The string to fix. */
	s string
	/* If this is true then don't change the string any further, jump
	   over it. */
	changed bool
}

type fixstrings []fixstring

func (in fixstrings) Join() (s string) {
	for i := range in {
		s += in[i].s
	}
	return s
}

var url_comp = "(?:[a-zA-Z0-9]+)"

// A query cannot not contain quotation marks.
var nobad = `[^\s"]*`

var query = "(?:" + nobad + "[0-9a-zA-Z/\\)])"
var url_regex = "https?://(?:" + url_comp + "\\.)*" + url_comp +
	"(?::[0-9]+)?" + "(?:/" + query + ")"
var url_replace = regexp.MustCompile(url_regex)

func urlToLink(from []byte) []byte {
	var url = string(from)
	var link = fmt.Sprintf("<a target='_blank' href='%s'>%s</a>", url, url)
	return []byte(link)
}

// Regular expression to match

var bugRegex = "[bB]ug\\s+([0-9]+)"
var bugReplace = regexp.MustCompile(bugRegex)

// Replacement text

var urlFmt = "<a target='_blank' href='%s/bug/%s'>%s</a>"

// Replace instances of "bug n" with a link.
func (b *Bagreply) replaceBugN(ins fixstrings) (outs fixstrings) {
	outs = make([]fixstring, 0)
	for i := range ins {
		if ins[i].changed {
			outs = append(outs, ins[i])
			continue
		}
		in := ins[i].s

		// The -1 at the end tells the finder that it doesn't need to
		// limit the number of results.
		var results = bugReplace.FindAllStringSubmatchIndex(in, -1)
		if len(results) == 0 {
			outs = append(outs, ins[i])
			continue
		}
		var end = 0
		for num, r := range results {
			if r[0] > end {
				before := fixstring{s: in[end:r[0]], changed: false}
				outs = append(outs, before)
			}
			bugText := in[r[0]:r[1]]
			bugNumber := in[r[2]:r[3]]
			out := fmt.Sprintf(urlFmt, b.App.TopURL, bugNumber, bugText)
			match := fixstring{s: out, changed: true}
			outs = append(outs, match)
			end = r[1]
			if num == len(results)-1 {
				after := fixstring{s: in[end:], changed: false}
				outs = append(outs, after)
			}
		}
	}
	return outs
}

func (b *Bagreply) bugsToLinks(ins fixstrings) (outs fixstrings) {
	outs = b.replaceBugN(ins)
	return outs
}

// Replace URLs in the text with actual links
func (b *Bagreply) urlsToLinks(input string) (out string) {
	ins := fixstrings{fixstring{s: input, changed: false}}
	outs := make(fixstrings, 0)
	for i := range ins {
		if ins[i].changed {
			outs = append(outs, ins[i])
			continue
		}
		in := ins[i].s
		urls := url_replace.FindAllStringIndex(in, -1)
		start := 0
		for _, url := range urls {
			startUrl := url[0]
			endUrl := url[1]
			if startUrl > start {
				outs = append(outs, fixstring{s: html.EscapeString(in[start:startUrl]), changed: false})
			}
			s := string(urlToLink([]byte(in[startUrl:endUrl])))
			outs = append(outs, fixstring{s: s, changed: true})
			start = endUrl
		}
		// Get the bit after the final URL
		if start < len(input) {
			outs = append(outs, fixstring{s: html.EscapeString(input[start:len(input)]), changed: false})
		}
	}
	outs = b.bugsToLinks(outs)
	out = outs.Join()
	return out
}
