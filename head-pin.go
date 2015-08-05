package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"image"
	"image/gif"
	"os"
	"bytes"
	"sync"
	"math"
)

type Frame struct {
    zoom int32
    frame  image.Image
    err  error
}

func gen() <-chan Frame {
	var (
		framesQty float64 = 10
		maxZoom float64 = 21
		minZoom float64 = 2
	)
	
	step := math.Ceil((maxZoom - minZoom) / framesQty)
	step = 1
	
	out := make(chan Frame)
    go func() {
    	for i := maxZoom; i >= minZoom; i -= step {
    		var f = Frame{int32(i), nil, nil}
    		out <- f
    	}
    	close(out)
    }()
    return out
}

func merge(cs ...<-chan Frame) <-chan Frame {
    var wg sync.WaitGroup
    out := make(chan Frame)

    // Start an output goroutine for each input channel in cs.  output
    // copies values from c to out until c is closed, then calls wg.Done.
    output := func(c <-chan Frame) {
        for n := range c {
            out <- n
        }
        wg.Done()
    }
    wg.Add(len(cs))
    for _, c := range cs {
        go output(c)
    }

    // Start a goroutine to close out once all the output goroutines are
    // done.  This must start after the wg.Add call.
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
	in := gen()

	pics1 := getFrame(in)
	pics2 := getFrame(in)
	pics3 := getFrame(in)
	pics4 := getFrame(in)
	pics5 := getFrame(in)
	pics6 := getFrame(in)
	pics7 := getFrame(in)
	pics8 := getFrame(in)
	pics9 := getFrame(in)
	pics0 := getFrame(in)
	pics11 := getFrame(in)
	pics22 := getFrame(in)
	pics33 := getFrame(in)
	pics44 := getFrame(in)
	pics55 := getFrame(in)
	pics66 := getFrame(in)
	pics77 := getFrame(in)
	pics88 := getFrame(in)
	pics99 := getFrame(in)
	pics00 := getFrame(in)

	pics100 := merge(pics1, pics2, pics3, pics4, pics5)
	pics200 := merge(pics6, pics7, pics8, pics9, pics0)
	pics300 := merge(pics11, pics22, pics33, pics44, pics55)
	pics400 := merge(pics66, pics77, pics88, pics99, pics00)
	
	pics := merge(pics100, pics200, pics300, pics400)

	var frames []*image.Paletted
	var delays []int

	for f := range pics {

		frames = append(frames, f.frame.(*image.Paletted))
		delays = append(delays, 10)
	}

	// ok, write out the data into the new GIF file
	out, err := os.Create("./output.gif")
	
	if err != nil {
		fmt.Println(err)
	    os.Exit(1)
	}

	opts := gif.GIF{
		frames,
		delays,
		1,
	}

	err = gif.EncodeAll(out, &opts)
	if err != nil {
		fmt.Println(err)
	    os.Exit(1)
	}
	
	fmt.Println("Generated image to output.gif \n")
}

func getFrame(in <-chan Frame) <-chan Frame {
	out := make(chan Frame)

	go func() {

		for f := range in {
			
			zoom := fmt.Sprintf("%d", f.zoom)
			
			url := "https://maps.googleapis.com/maps/api/staticmap?center=Brooklyn%20Bridge%2CNew%20York%2CNY&zoom="+zoom+"&size=200x200&maptype=satellite&key=AIzaSyDmg9lTp4x64M08Mr4YWPKPN0ZO0BUp88U&format=gif"

			req, _ := http.NewRequest("GET", url, nil)

			res, _ := http.DefaultClient.Do(req)

			defer res.Body.Close()
			
			body, _ := ioutil.ReadAll(res.Body)

			// convert []byte to image for saving to file
		   	img, _, err := image.Decode(bytes.NewReader(body))
		   	
		   	f.frame = img
			f.err = err
			out <- f
    	}

    	close(out)
	}()

	return out
}