package helper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type ImageOp struct {
	Width                                   int
	Height                                  int
	Crop                                    bool
	Embed					bool
	Filename                                string
	Hash                                    string
	Format, Top, Left, X, Y, Angle, Quality int
	Gravity                                 string
}

// do all magic behind parsing url and return struct
func Prepare(surl string) ImageOp {
	var (
		width, height, format, top, left, x, y, a int
		gravity                             string
		err                                       error
		q                                         int = 85
		crop, embed bool
	)

	// split url to parts
	parts := strings.SplitN(surl, "/", 5)

	// get file name, and resize string
	filename, _ := url.QueryUnescape(parts[4])
	commands := parts[3]

	// regexp for width and height
	rwh := regexp.MustCompile("([wh])+_(\\d+)")

	// extract
	ret := regexp.MustCompile("([tlxy])+_(\\d+)")

	// regexp for crop
	rcr := regexp.MustCompile("c_(\\w+)")

	// regexp for embed
	rem := regexp.MustCompile("e_(\\w+)")

	// regexp for gravity
	rgr := regexp.MustCompile("g_(\\w+)")

	// regexp for type
	rfc := regexp.MustCompile("t_(\\w+)")

	// regexp for angle
	ran := regexp.MustCompile("a_(\\d+)")

	// regexp for quality
	rqa := regexp.MustCompile("q_(\\d+)")

	// split all command to smaller sections
	z := strings.Split(commands, ",")

	for _, value := range z {

		// match width and height
		if rwh.MatchString(value) {
			matches := rwh.FindStringSubmatch(value)

			if matches[1] == "w" {
				width, err = strconv.Atoi(matches[2])
			}
			if matches[1] == "h" {
				height, err = strconv.Atoi(matches[2])
			}
		}

		// match extract
		if ret.MatchString(value) {
			matches := ret.FindStringSubmatch(value)

			if matches[1] == "t" {
				top, err = strconv.Atoi(matches[2])
			}
			if matches[1] == "l" {
				left, err = strconv.Atoi(matches[2])
			}
			if matches[1] == "x" {
				x, err = strconv.Atoi(matches[2])
			}
			if matches[1] == "y" {
				y, err = strconv.Atoi(matches[2])
			}
		}

		// angle (rotate)
		if ran.MatchString(value) {
			matches := ran.FindStringSubmatch(value)

			a, err = strconv.Atoi(matches[1])

		}

		// quality
		if rqa.MatchString(value) {
			matches := rqa.FindStringSubmatch(value)
			q, err = strconv.Atoi(matches[1])

			if q == 0 || q > 100 {
				q = 85
			}
		}

		// crop
		if rcr.MatchString(value) {
			matches := rcr.FindStringSubmatch(value)
			if matches[1] == "true" {
				crop = true
			} else {
				crop = false
			}
		}
		// embed
		if rem.MatchString(value) {
			matches := rem.FindStringSubmatch(value)
			if matches[1] == "true" {
				embed = true
			} else {
				embed = false
			}
		}


		// gravity
		if rgr.MatchString(value) {
			matches := rgr.FindStringSubmatch(value)
			gravity = matches[1]
		}

		// type conversion
		if rfc.MatchString(value) {
			matches := rfc.FindStringSubmatch(value)

			switch matches[1] {
			case "jpg":
				format = 1
			case "webp":
				format = 2
			case "png":
				format = 3
			case "tiff":
				format = 4
			}
		}
	}

	// check if we have correct integer values
	if err != nil {
		fmt.Println("wrong width or height")
	}

	return ImageOp{width, height, crop, embed, filename, GetMD5Hash(filename + commands), format, top, left, x, y, a, q, gravity}
}

func GetImageType(ext string) string {
	switch ext {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"

	case "tiff":
		return "image/tiff"
	}
	return ""
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetMD5HashB(text []byte) string {
	hasher := md5.New()
	hasher.Write(text)
	return hex.EncodeToString(hasher.Sum(nil))
}
