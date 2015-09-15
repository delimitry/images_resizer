//-----------------------------------------------------------------------------
// Author: delimitry
//-----------------------------------------------------------------------------

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Average RGBA color from 2x2 rect
func average2x2(c00, c01, c10, c11 color.Color) color.Color {
	c00r, c00g, c00b, c00a := c00.RGBA()
	c01r, c01g, c01b, c01a := c01.RGBA()
	c10r, c10g, c10b, c10a := c10.RGBA()
	c11r, c11g, c11b, c11a := c11.RGBA()
	// Each RGBA channel is in the range [0..65535]
	c := color.RGBA{
		uint8(float32(c00r+c01r+c10r+c11r) / 1024.0),
		uint8(float32(c00g+c01g+c10g+c11g) / 1024.0),
		uint8(float32(c00b+c01b+c10b+c11b) / 1024.0),
		uint8(float32(c00a+c01a+c10a+c11a) / 1024.0)}
	return c
}

// Clamp x and y to [0..w-1] and [0..h-1] respectively
func clamp(x, y, w, h int) (int, int) {
	if x < 0 {
		x = 0
	} else if x > w-1 {
		x = w - 1
	}
	if y < 0 {
		y = 0
	} else if y > h-1 {
		y = h - 1
	}
	return x, y
}

// Resize image
func resizeImage(img image.Image, factor float64) image.Image {
	out := image.NewRGBA(image.Rect(0, 0,
		int(float64(img.Bounds().Dx())*factor), int(float64(img.Bounds().Dy())*factor)))
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			x := int(float64(j) * factor)
			y := int(float64(i) * factor)
			c := average2x2(
				img.At(clamp(j, i, w, h)), img.At(clamp(j+1, i, w, h)),
				img.At(clamp(j, i+1, w, h)), img.At(clamp(j+1, i+1, w, h)))
			out.Set(x, y, c)
		}
	}
	return out
}

// Resize images found in folder one by one
func resizeImages(path string, factor float64) {
	possibleExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		filename := f.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		name := f.Name()[0 : len(filename)-len(ext)]
		// skip already resized images
		if strings.Contains(name, "_resized") {
			continue
		}
		if possibleExts[ext] {
			fmt.Printf("Resizing %s\n", filename)

			// open image file
			file, err := os.Open(filepath.Join(path, filename))
			if err != nil {
				log.Fatal(err)
			}
			img, format, err := image.Decode(file)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			// resize
			resized := resizeImage(img, factor)

			// create out file
			out, err := os.Create(filepath.Join(path, name+"_resized"+ext))
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			// write to out according to format
			if format == "jpeg" {
				jpeg.Encode(out, resized, nil)
			} else if format == "png" {
				png.Encode(out, resized)
			} else if format == "gif" {
				gif.Encode(out, resized, nil)
			}
		}
	}
}

// Resize images concurrently - one resize func goroutine for image file
func asyncResizeImages(path string, factor float64) {
	possibleExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	var imageFiles []string
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		filename := f.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		name := filename[0 : len(filename)-len(ext)]
		// add only files with image ext and skip already resized
		if !strings.Contains(name, "_resized") && possibleExts[ext] {
			imageFiles = append(imageFiles, filename)
		}
	}
	results := make(chan string, len(imageFiles))
	for _, fn := range imageFiles {
		go func(path string, filename string, factor float64) {
			ext := strings.ToLower(filepath.Ext(filename))
			name := filename[0 : len(filename)-len(ext)]

			fmt.Printf("Resizing %s\n", filename)

			// open image file
			file, err := os.Open(filepath.Join(path, filename))
			if err != nil {
				log.Fatal(err)
			}
			img, format, err := image.Decode(file)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			// resize
			resized := resizeImage(img, factor)

			// create out file
			out, err := os.Create(filepath.Join(path, name+"_resized"+ext))
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			// write to out according to format
			if format == "jpeg" {
				jpeg.Encode(out, resized, nil)
			} else if format == "png" {
				png.Encode(out, resized)
			} else if format == "gif" {
				gif.Encode(out, resized, nil)
			}
			results <- filename
		}(path, fn, factor)
	}

	// collect the results
	for i := 0; i < len(imageFiles); i++ {
		fmt.Printf("%s OK [%d]\n", <-results, i)
	}
}

// Resize images using pool of workers to run several concurrent resizer funcs
func workersPoolResizeImages(path string, factor float64) {
	possibleExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	var imageFiles []string
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		filename := f.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		name := filename[0 : len(filename)-len(ext)]
		// add only files with image ext and skip already resized
		if !strings.Contains(name, "_resized") && possibleExts[ext] {
			imageFiles = append(imageFiles, filename)
		}
	}

	const workersNum int = 5
	inputs := make(chan string, len(imageFiles))
	results := make(chan string, len(imageFiles))

	for i := 0; i < workersNum; i++ {
		go func(path string, factor float64, i int, inputs <-chan string, results chan<- string) {
			for filename := range inputs {
				ext := strings.ToLower(filepath.Ext(filename))
				name := filename[0 : len(filename)-len(ext)]
				fmt.Printf("Resizing %s (worker %d)\n", filename, i)

				// open image file
				file, err := os.Open(filepath.Join(path, filename))
				if err != nil {
					log.Fatal(err)
				}
				img, format, err := image.Decode(file)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				// resize
				resized := resizeImage(img, factor)

				// create out file
				out, err := os.Create(filepath.Join(path, name+"_resized"+ext))
				if err != nil {
					log.Fatal(err)
				}
				defer out.Close()

				// write to out according to format
				if format == "jpeg" {
					jpeg.Encode(out, resized, nil)
				} else if format == "png" {
					png.Encode(out, resized)
				} else if format == "gif" {
					gif.Encode(out, resized, nil)
				}
				results <- filename
			}
		}(path, factor, i, inputs, results)
	}

	// send filenames to inputs channel
	for _, fn := range imageFiles {
		inputs <- fn
	}
	close(inputs)

	// collect the results
	for i := 0; i < len(imageFiles); i++ {
		fmt.Printf("%s OK [%d]\n", <-results, i)
	}
}

func main() {
	var dirWithImages = flag.String("d", ".", "Directory with images to resize")
	var scalingFactor = flag.Float64("f", 0.5, "Scaling factor value from (0.0 to 1.0)")

	flag.Usage = func() {
		fmt.Printf("Images resizer tool by Delimitry (c) 2015\n" +
			"Usage: images_resizer [options]\n\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	if (len(os.Args) < 3) || (*scalingFactor <= 0.0 || *scalingFactor >= 1.0) {
		flag.Usage()
	}

	start := time.Now()
	fmt.Println(strings.Repeat("-", 80))
	workersPoolResizeImages(*dirWithImages, *scalingFactor)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Resize images time: %s\n", time.Since(start))
}
