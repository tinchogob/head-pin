package main

import (
    "fmt"
    "net/http"
    "os"
    "golang.org/x/net/context"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/urlfetch"
	"io"
    "io/ioutil"
	"image"
	"image/gif"
	"bytes"
	"sync"
	"math"
	"sort"
	"strconv"
)

func getLogger(c context.Context) func(msg string) {
	return func(msg string) {
		log.Debugf(c, msg)
	}
}

func init() {
    http.HandleFunc("/pins/me", pins_me)
    http.HandleFunc("/ping", health)
}

func health(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "pong")
}

func pins_me(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// logger := getLogger(c)
	if r.Method == "GET" {
		http.ServeFile(w, r, "./index.html")
	} else if r.Method == "POST" {
		lat := r.PostFormValue("lat")
		lon := r.PostFormValue("lon")
		workers, _ := strconv.Atoi(r.PostFormValue("workers"))
		delay, _ := strconv.Atoi(r.PostFormValue("delay"))
		loops, _ := strconv.Atoi(r.PostFormValue("loops"))
		alto := r.PostFormValue("alto")
		ancho := r.PostFormValue("ancho")
		zoomType := r.PostFormValue("zoomType")

		pin := PIN{lat, lon, workers, delay, loops, alto, ancho, zoomType}
		// pin := PIN{"-34.5667166","-58.4574169", 10, 100, 20, "500", "800"}

		Headpin(pin, w, c);
	}

}

func main() {
/*
	out, err := os.Create("out.gif")

	if err != nil {
		fmt.Println("Error creando el archivo")
		return
	}

    Headpin(10, 100, 1, out);
*/
    os.Exit(0)
}

type Frame struct {
    zoom int
    frame  image.Image
    err  error
}

func gen() <-chan Frame {
	var (
		framesQty float64 = 10
		maxZoom float64 = 20
		minZoom float64 = 3
	)
	
	step := math.Ceil((maxZoom - minZoom) / framesQty)
	step = 1
	
	out := make(chan Frame)
    go func() {
    	for i := maxZoom; i >= minZoom; i -= step {
    		var f = Frame{int(i), nil, nil}
    		out <- f
    	}
    	close(out)
    }()
    return out
}

func merge(processors []<-chan Frame) <-chan Frame {
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
    wg.Add(len(processors))
    for _, c := range processors {
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

// ByZoomIn implements sort.Interface for []Frame based on
// the Zoom field from space to ground.
type ByZoomIn []Frame

func (a ByZoomIn) Len() int           { return len(a) }
func (a ByZoomIn) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByZoomIn) Less(i, j int) bool { return a[i].zoom < a[j].zoom }

// ByZoomOut implements sort.Interface for []Frame based on
// the Zoom field from ground to space.
type ByZoomOut []Frame

func (a ByZoomOut) Len() int           { return len(a) }
func (a ByZoomOut) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByZoomOut) Less(i, j int) bool { return a[i].zoom > a[j].zoom }

type PIN struct {
	lat, lon string
	workers, delay, loops int
	alto, ancho, zoomType string
}

func Headpin(pin PIN, writer io.Writer, ctx context.Context) {
	// logger := getLogger(ctx)
	in := gen()

	var processors []<-chan Frame

	for i := 1; i <= pin.workers; i++ {
    	processors = append(processors, getFrame(pin, in, ctx))
	}

	pics := merge(processors)

	var frames []Frame
	var images []*image.Paletted
	var delays []int

	for f := range pics {
		if f.err != nil {
			continue
		}
		frames = append(frames, f)
	}

	if pin.zoomType == "in" {
		sort.Sort(ByZoomIn(frames))
	} else {
		sort.Sort(ByZoomOut(frames))
	}

	for i := 0; i < len(frames); i++ {
		f := frames[i]
		images = append(images, f.frame.(*image.Paletted))
		delays = append(delays, pin.delay)
	}

	for i := len(frames)-2; i >= 0 ; i-- {
		f := frames[i]
		images = append(images, f.frame.(*image.Paletted))
		delays = append(delays, pin.delay)
	}

	opts := gif.GIF{
		Image: images,
		Delay: delays,
		LoopCount: pin.loops,
	}

	err := gif.EncodeAll(writer, &opts)

	if err != nil {
		fmt.Println(err)
	    return
	}
}

func getFrame(pin PIN, in <-chan Frame, ctx context.Context) <-chan Frame {
	logger := getLogger(ctx)
	out := make(chan Frame)

	go func() {

		for f := range in {
			
			zoom := fmt.Sprintf("%d", f.zoom)
			
			url := "http://maps.googleapis.com/maps/api/staticmap?center="+pin.lat+","+pin.lon+"&zoom="+zoom+"&size="+pin.ancho+"x"+pin.alto+"&maptype=satellite&key=AIzaSyDmg9lTp4x64M08Mr4YWPKPN0ZO0BUp88U&format=gif"
			req, _ := http.NewRequest("GET", url, nil)
			
			client := urlfetch.Client(ctx)

			res, err := client.Do(req)

			logger(url)

			if err != nil {
				logger("Error al volver el request")
				logger(fmt.Sprintf("%v", err))
				f.err = err
				out <- f
				return
			} 

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
