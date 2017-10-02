package image

import (
	"fmt"
	//"github.com/slav123/bimg"
	//"gopkg.in/h2non/bimg.v0"
	"gopkg.in/h2non/bimg.v1"
	"io/ioutil"
	"log"
	"os"
	"resizer/helper"
	"time"
)

var LastError string

func DetermineImageTypeName(image *[]byte) string {

	bytes := make([]byte, 4, 4)
	copy(bytes, *image)

	if len(bytes) < 4 {
		return ""
	}

	if bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47 {
		return "png"
	}
	if bytes[0] == 0xFF && bytes[1] == 0xD8 {
		return "jpg"
	}
	if bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x38 {
		return "gif"
	}
	if bytes[0] == 0x42 && bytes[1] == 0x4D {
		return "bmp"
	}
	if bytes[0] == 0x52 && bytes[1] == 0x49 {
		return "webp"
	}
	return ""
}

/**
 * resize image action
 *
 * @param inBuf []byte image to resize
 * @param w, h int width and height
 * @param c
 */
func Resize(inBuf []byte, opt helper.ImageOp) []byte {

	options := bimg.Options{
		Width:  opt.Width,
		Height: opt.Height,
		//Crop:   c,
		//Embed : true,
		Interpolator: bimg.Bicubic,
		Gravity:      bimg.GravityCentre,
		Quality:      opt.Quality,
	}

	if opt.Gravity != "" {
		switch opt.Gravity {
		case "north":
			options.Gravity = bimg.GravityNorth
		case "south":
			options.Gravity = bimg.GravitySouth
		}

	}

	if opt.Crop  {
		options.Crop = true
	}

	if opt.Embed {
		options.Embed = true
	}

	log.Printf("crop: %t, embed: %t", opt.Crop, opt.Embed)

	// change image type
	if opt.Format > 0 {
		options.Type = bimg.ImageType(opt.Format)
	}

	// rotate
	if opt.Angle > 0 {
		switch opt.Angle {
		case 90:
			options.Rotate = bimg.D90
		case 180:
			options.Rotate = bimg.D180
		case 270:
			options.Rotate = bimg.D270
		}

	}

	image := bimg.NewImage(inBuf)

	// extract area
	if opt.Top > 0 || opt.Left > 0 {
		/*
			options.Top = top
			options.Left = left
			options.AreaWidth = aw
			options.AreaHeight = ah
		*/

		_, err := image.Extract(opt.Top, opt.Left, opt.X, opt.Y)
		if err != nil {
			fmt.Println("failed to extract image")
		}

		/*
			data, err := inBuf.Metadata()
			fmt.Printf("%v", data)
		*/

		//image = bimg.NewImage(temp)

	}

	newImage, err := image.Process(options)

	if err != nil {
		fmt.Println("failed to process image")
		log.Printf("options: %v", options)

		/*bimg.VipsDebugInfo()
		bimg.Shutdown()
		bimg.Initialize()
		*/
		fmt.Fprintln(os.Stderr, err)
		//LastError = err.Error()
		t := time.Now()
		fname := t.Format(time.RFC3339) + "." + DetermineImageTypeName(&inBuf)

		ioutil.WriteFile("cache/"+fname, inBuf, 0644)

	}

	return newImage
}

func Convert(inBuf []byte, t int) []byte {

	var it bimg.ImageType

	it = bimg.ImageType(t)

	newImage, err := bimg.NewImage(inBuf).Convert(it)

	if err != nil {
		fmt.Println("failed to convert image")
		fmt.Fprintln(os.Stderr, err)
	}

	return newImage
}
