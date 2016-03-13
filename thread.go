// API help for 4chan.
package fourchan

/*
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

// Meta information about a post in a thread.
// Note that some fields are optional and may contain only their default values.
// https://github.com/4chan/4chan-API
type Meta struct {
	// Thread is archived?
	Archived bool
	// Max number of bumps?
	BumpLimit bool
	// Thread closed?
	Closed bool
	// Image was deleted?
	FileDeleted bool
	// Synthesized, has an image?
	HasFile bool
	// Has reached image limit?
	ImageLimit bool
	// Is spoiler post?
	Spoiler bool
	// Is sticky post?
	Sticky bool

	// The ID for this post
	PostNumber uint64 `json:"no"`
	// The ID this replies to (0 is OP, what if more than one?)
	ReplyTo uint64 `json:"resto"`

	// Seconds since epoch
	UnixTime uint64 `json:"time"`
	// unix time last modified
	LastModified uint64 `json:"last_modified"`
	// String based time representation
	Time string `json:"now"`

	// admin, mod, etc
	AdminId string `json:"id"`
	// admin, mod, etc
	AdminType string `json:"capcode"`
	// Which comments in the thread are admin replies (I think?)
	AdminReplies []uint64 `json:"admin"`

	// Look at me look at me
	Name string `json:"name"`
	// Insufferable moron identification
	TripCode string `json:"trip"`

	// Poster's country code
	CountryCode string `json:"country"`
	// Poster's country
	Country string `json:"country_name"`

	// The original filename
	OrigFileName string `json:"filename"`
	// The file type (.jpeg, etc)
	FileExt string `json:"ext"`
	// The new name of the file [0-9]+.ext
	RenamedFileName uint64 `json:"tim"`

	// md5sum of the image
	FileMD5 string `json:"md5"`
	// Size of the image
	FileSize int `json:"fsize"`
	// Height of image for this post
	FileHeight int `json:"h"`
	// Width of image for this post
	FileWidth int `json:"w"`

	// Height of thumbnail for this post's image (when?)
	ThumbnailHeight int `json:"tn_h"`
	// Width of thumbnail for this post's image (when?)
	ThumbnailWidth int `json:"tn_w"`

	// The id of the custom spoiler image
	CustomSpoiler int `json:"custom_spoiler"`
	// Number of posts not in this object (when does this happen?)
	OmittedPosts int `json:"omitted_posts"`
	// Number of images not in this object (when does this happen?)
	OmittedImages int `json:"omitted_images"`

	// Number of images in thread (does this actually show up for this response?)
	ReplyCount int `json:"replies"`
	// Number of images in thread (does this actually show up for this response?)
	ImageCount int `json:"images"`

	// ???
	Tag string `json:"tag"`
	// Only occurs at the top level of the post
	SemanticUrl string `json:"semantic_url"`
}

// A single post in a thread.
// Note that some fields are optional and may contain only their default values.
// https://github.com/4chan/4chan-API
type Post struct {
	// The post subject
	Subject string `json:"sub"`
	// The text of the post
	Comment string `json:"com"`

	// OrigFileName + . + FileExt
	FullOrigFileName string

	// str(RenamedFileName) + . + FileExt
	FullNewFileName string

	// All of the meta info for this post
	Meta
}

// Converts integer values to boolean values
// 0 = false all other values = true
func intToBool(i int) bool {
	if i == 0 {
		return false
	}
	return true
}

// Converts bool values to int values
// false = 0, true = 1
func boolToInt(b bool) int {
	if !b {
		return 0
	}
	return 1
}

// Custom marshaler for a Post struct.
// We have to handle the conversion from bools to ints :(
func (p *Post) MarshalJSON() ([]byte, error) {
	type Alias Post
	return json.Marshal(&struct {
		*Alias

		ArchivedInt    int `json:"archived"`
		BumpLimitInt   int `json:"bumplimit"`
		ClosedInt      int `json:"closed"`
		FileDeletedInt int `json:"filedeleted"`
		ImageLimitInt  int `json:"imagelimit"`
		SpoilerInt     int `json:"spoiler"`
		StickyInt      int `json:"sticky"`
	}{
		Alias: (*Alias)(p),

		ArchivedInt:    boolToInt(p.Archived),
		BumpLimitInt:   boolToInt(p.BumpLimit),
		ClosedInt:      boolToInt(p.Closed),
		FileDeletedInt: boolToInt(p.FileDeleted),
		ImageLimitInt:  boolToInt(p.ImageLimit),
		SpoilerInt:     boolToInt(p.Spoiler),
		StickyInt:      boolToInt(p.Sticky),
	})
}

// Custom marshaler for a Post struct.
// We have to handle the conversion from ints to bools :(
func (p *Post) UnmarshalJSON(data []byte) error {
	type Alias Post
	tmp := &struct {
		*Alias

		ArchivedInt    int `json:"archived"`
		BumpLimitInt   int `json:"bumplimit"`
		ClosedInt      int `json:"closed"`
		FileDeletedInt int `json:"filedeleted"`
		ImageLimitInt  int `json:"imagelimit"`
		SpoilerInt     int `json:"spoiler"`
		StickyInt      int `json:"sticky"`
	}{
		Alias: (*Alias)(p),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	p.Archived = intToBool(tmp.ArchivedInt)
	p.BumpLimit = intToBool(tmp.BumpLimitInt)
	p.Closed = intToBool(tmp.ClosedInt)
	p.FileDeleted = intToBool(tmp.FileDeletedInt)
	p.ImageLimit = intToBool(tmp.ImageLimitInt)
	p.Spoiler = intToBool(tmp.SpoilerInt)
	p.Sticky = intToBool(tmp.StickyInt)

	p.FullOrigFileName = p.OrigFileName + p.FileExt
	if p.RenamedFileName != 0 {
		p.FullNewFileName = strconv.FormatUint(p.RenamedFileName, 10) + p.FileExt
		p.HasFile = true
	}

	return nil
}

// A thread.
// We add the board to this to ease the work of interface consumers.
type Thread struct {
	// The list of comments in this thread.
	Posts []Post `json:"posts"`
	// The board this thread is on.
	Board string
}

// Custom error to indicate we were unable to extract necessary info from the provided URL.
type URLMatchError struct {
	url string
}

// Pretty print dat error yo.
func (e URLMatchError) Error() string {
	return fmt.Sprintf("Could not extract thread info from %s", e.url)
}

// Extract the board and thread ID from a given URL.
func extractBoardAndThreadId(url string) (board string, id string, err error) {
	err = nil
	board = ""
	id = ""

	reg, err := regexp.Compile("https?://[^./]*\\.4[^./]*\\.org/([^/]*)/thread/([0-9]*)(?:(?:/|#).*)?")
	if err != nil {
		return
	}
	matches := reg.FindAllStringSubmatch(url, -1)

	if len(matches) != 1 || len(matches[0]) != 3 {
		err = URLMatchError{url}
		return
	}

	board = matches[0][1]
	id = matches[0][2]

	return
}

// Given an URL, extract the board and thread ID then load the thread.
func LoadThreadFromURL(url string) (*Thread, error) {
	board, id, err := extractBoardAndThreadId(url)
	if err != nil {
		return nil, err
	}

	return LoadThreadById(board, id)
}

// Load a thread by board and ID.
func LoadThreadById(board, id string) (*Thread, error) {
	url := fmt.Sprintf("https://a.4cdn.org/%s/thread/%s.json", board, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	thread := &Thread{}
	err = json.Unmarshal(bodyBytes, thread)
	if err != nil {
		return nil, err
	}

	thread.Board = board

	return thread, nil
}
