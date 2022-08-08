package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
)

func random(max int64) int {
	r, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return int(r.Int64())
}

func main() {
	mb := 1 + random(255)
	if mmb, ok := os.LookupEnv("HTML_MB_SIZE"); ok {
		mmb, err := strconv.Atoi(mmb)
		if err == nil {
			mb = mmb
		}
	}
	port := ":80"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = p

	}
	http.HandleFunc("/", CreateHandler(mb*1024*1024))
	fmt.Printf("serving %d mb random html output.\n", mb)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}

func paragraph() []byte {
	return []byte(fmt.Sprintf("<p>%x</p>", random(9999)))
}

func CreateHandler(wanted_size int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write([]byte("<html>"))
		w.Write([]byte("<body>"))
		flusher.Flush()
		fmt.Printf("generating until %db is reached.\n", wanted_size)
		sb := 0
		for sb < wanted_size {
			//fmt.Printf("\rgenerated %d / %db", sb, wanted_size)
			p := paragraph()
			sb = sb + len(p)
			w.Write(p)
			flusher.Flush()
		}
		w.Write([]byte("</body>"))
		w.Write([]byte("</html>"))
		flusher.Flush()
	}

}
