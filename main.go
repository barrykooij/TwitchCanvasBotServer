package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"image"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Pixel struct {
	X, Y  int
	Color string
}

var pixels []*Pixel

var baseX int = 0
var baseY int = 0

var deck []*Pixel

func CORSHandler(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// JSON content type
		w.Header().Set("Content-Type", "application/json")

		// for now we allow origin from give domain
		//domain := r.Header.Get("Origin")
		//w.Header().Set("Access-Control-Allow-Origin", domain)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"DNT, Access-Control-Allow-Headers, Origin, X-Requested-With,  Access-Control-Request-Method, Access-Control-Request-Headers, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}

		// ServeHTTP
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
func loadPixels(file io.Reader) {
	img, _, err := image.Decode(file)

	if err != nil {
		panic(err.Error())
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	//var pixels []Pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			if a == 0 {
				continue
			}

			pixels = append(pixels, &Pixel{
				X: x + baseX,
				Y: y + baseY,
				//Color: fmt.Printf("%02x%02x%02x", r, g, b),
				Color: strings.ToUpper(fmt.Sprintf("%02x%02x%02x", int(r/257), int(g/257), int(b/257))),
			})
			//fmt.Printf("Alpha: %d\n", int(a/257))
		}
		//pixels = append(pixels, row)
	}
}

func createDeck() {
	deck = make([]*Pixel, len(pixels))
	copy(deck, pixels)
	// shuffle new deck
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
}

func getNextFromDeck() *Pixel {
	pixel := deck[0] // get the 0 index element from slice
	deck = deck[1:]  // remove the 0 index element from slice
	return pixel
}

func loadImage() {

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	file, err := os.Open("./image.png")

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	loadPixels(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

}

/*
var currentPixel *Pixel
func getNextVoidPixel() *Pixel {
	p := currentPixel

	a := -1
	b := 1

	p.X = p.X + (a + rand.Intn(b-a+1))
	p.Y = p.Y + (a + rand.Intn(b-a+1))

	return p
}
*/

func pixelRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	if len(deck) == 0 {
		createDeck()
	}

	p := getNextFromDeck()

	j, e := json.Marshal(p)
	if e != nil {
		panic(e.Error())
	}

	fmt.Printf("Placing %d, %d, %s\n", p.X, p.Y, p.Color)

	_, e = fmt.Fprintf(w, string(j))
	if e != nil {
		fmt.Printf("Error: %s", e.Error())
	}
}

func main() {

	flag.IntVar(&baseX, "x", 0, "start position X")
	flag.IntVar(&baseY, "y", 0, "start position Y")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	loadImage()

	deck = make([]*Pixel, 0)

	fmt.Printf("Pixels found: %d", len(pixels))

	chain := alice.New(CORSHandler)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Why are you here?")
	})

	r.Handle("/Pixel", chain.ThenFunc(pixelRoute))
	http.Handle("/", r)

	//log.Fatal(http.ListenAndServe("localhost:8080", nil))
	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/ttvcanvashelper.fun/fullchain.pem", "/etc/letsencrypt/live/ttvcanvashelper.fun/privkey.pem", nil))
}
