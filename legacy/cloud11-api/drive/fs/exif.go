package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/ihleven/cloud11-api/drive"
)

func exifDecode(i *drive.Image, fd *os.File) error {
	// https://github.com/rwcarlsen/goexif

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	//exif.RegisterParsers(mknote.All...)

	fmt.Println("exifDecode", i, fd)
	i.Exif = &drive.Exif{}

	fd.Seek(0, 0)
	x, err := exif.Decode(fd)
	if err != nil {
		return err
	}

	camModel, err := x.Get(exif.Model) // normally, don't ignore errors!
	if err != nil {
		return err
	}
	model, err := camModel.StringVal()
	if err != nil {
		return err
	}
	i.Exif.Model = model

	orientation, err := x.Get(exif.Orientation)
	if err != nil {
		return err
	}
	o, err := orientation.Int(0)

	if err != nil {
		return err
	}
	i.Exif.Orientation = float64(o)

	focal, _ := x.Get(exif.FocalLength)
	numer, denom, err := focal.Rat2(0) // retrieve first (only) rat. value
	if err != nil {
		return err
	}
	fmt.Printf("%v/%v %s\n", numer, denom, focal.String())

	// Two convenience functions exist for date/time taken and GPS coords:
	_, err = x.DateTime()
	if err != nil {
		return err
	}
	//i.Exif.Taken = &tm

	lat, long, err := x.LatLong()
	if err != nil {
		return err
	}
	i.Exif.GPSLatitude = lat
	i.Exif.GPSLongitude = long

	//j := x.String()
	//fmt.Printf("json: %s", j)

	return nil
}

func metaFilename(path string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return fmt.Sprintf("%s.txt", base)
}

func parseMeta(path string, i *drive.Image) error {

	re := regexp.MustCompile(`(?s)(?P<Title>.*?)=+(?P<Caption>.*?)---+(?P<Cutline>.*?)---+`)

	fd, err := os.Open(metaFilename(path))
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	match := re.FindSubmatch(content)
	paramsMap := make(map[string]string)

	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = strings.TrimSpace(string(match[i]))
		}
	}
	if title, ok := paramsMap["Title"]; ok {
		i.Title = title
	}
	if caption, ok := paramsMap["Caption"]; ok {
		i.Caption = caption
	}
	if cutline, ok := paramsMap["Cutline"]; ok {
		i.Cutline = cutline
	}
	return nil
}
