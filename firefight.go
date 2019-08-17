package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"goji.io"
	"goji.io/pat"

	"firefight"
)

var ffServers sync.Map

func loadServer(id string) *firefight.FireFight {
	ffi, ok := ffServers.Load(id)
	if ok {
		return ffi.(*firefight.FireFight)
	}

	ffi, loaded := ffServers.LoadOrStore(id, firefight.New())
	if !loaded {
		log.Printf("[ffserver][%s] Created.\n", id)
	}

	return ffi.(*firefight.FireFight)
}

func main() {
	// {
	// 	testIDs := []string{
	// 		"AAA",
	// 		"BBB",
	// 		"CCC",
	// 		"DDD",
	// 		"EEE",
	// 	}
	// 	ff := loadServer("DEBUG")
	// 	for _, id := range testIDs {
	// 		ff.Join(id)
	//
	// 	}
	// 	ff.Start()
	//
	// 	for _, id := range testIDs {
	// 		ff.ReportHit(id)
	// 	}
	//
	// 	ff.Pause()
	// }
	// {
	// 	testIDs := []string{
	// 		"AAA",
	// 		"BBB",
	// 		"CCC",
	// 		"DDD",
	// 		"EEE",
	// 	}
	// 	ff := loadServer("DEBUG2")
	// 	for _, id := range testIDs {
	// 		ff.Join(id)
	//
	// 	}
	// 	ff.Start()
	//
	// 	for _, id := range testIDs {
	// 		ff.ReportHit(id)
	// 	}
	//
	// 	ff.Pause()
	// }

	endpoint := goji.SubMux()
	endpoint.Use(Context)
	endpoint.Use(func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			scmd, ok := r.Context().Value("slack_cmd").(*firefight.SlackCmd)
			if !ok {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			defer func(start time.Time) {
				log.Printf("[ffserver][%s][%s][%s] request completed in: %s",
					scmd.ChannelID, scmd.Command, scmd.UserID, time.Since(start))
			}(time.Now())

			ff := loadServer(scmd.ChannelID)
			ctx := context.WithValue(r.Context(), "fire_fight", ff)

			h.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	})

	endpoint.HandleFunc(pat.Post("/ffstart"), firefight.Start)
	endpoint.HandleFunc(pat.Post("/ffpause"), firefight.Pause)
	endpoint.HandleFunc(pat.Post("/ffend"), firefight.End)

	endpoint.HandleFunc(pat.Post("/ffjoin"), firefight.Join)
	endpoint.HandleFunc(pat.Post("/fftarget"), firefight.Target)
	endpoint.HandleFunc(pat.Post("/ffhit"), firefight.ReportHit)
	endpoint.HandleFunc(pat.Post("/ffdispute"), firefight.DisputeHit)

	endpoint.HandleFunc(pat.Post("/ffscore"), firefight.Scoreboard)

	mux := goji.NewMux()
	mux.Handle(pat.New("/endpoint/*"), endpoint)
	mux.Handle(pat.New("/debug/*"), DebugRoutes())
	mux.HandleFunc(pat.Get("/"), Index)

	log.Fatal(http.ListenAndServe(":8081", mux))
}

func Context(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("[Context]", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		m, err := url.ParseQuery(string(body))
		if err != nil {
			log.Println("[Context]", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		cmd := firefight.ParseSlackCmd(m)
		ctx := context.WithValue(r.Context(), "slack_cmd", cmd)

		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

const helpText = `/endpoint/ffstart
/endpoint/ffpause
/endpoint/ffend
/endpoint/ffjoin
/endpoint/fftarget
/endpoint/ffhit
/endpoint/ffdispute
/endpoint/ffscore
`

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, helpText)
}

func DebugRoutes() *goji.Mux {
	debugMux := goji.SubMux()
	debugMux.HandleFunc(pat.Get("/"), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<!DOCTYPE html><html><head><title>FireFight Debug</title></head><body>")
		ffServers.Range(func(key, value interface{}) bool {
			fmt.Fprintf(w, "<p><a href='./channel/%s'>%s</a></p>\n", key, key)
			return true
		})
		fmt.Fprintln(w, "</body></html>")
	})

	debugMux.HandleFunc(pat.Get("/channel/:id"), func(w http.ResponseWriter, r *http.Request) {
		id := pat.Param(r, "id")
		ffi, ok := ffServers.Load(id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ffi); err != nil {
			log.Println("[Join]", err)
		}
	})

	return debugMux
}

// Slack endpoint verification.
//
// const Signing_Secret = ""
//
// const ChallengeTimeRange = 5 * time.Minute
//
// func SlackHeaderHMAC(h http.Header) []byte {
// 	sigHeader, err := url.ParseQuery(h.Get("X-Slack-Signature"))
// 	if err != nil {
// 		return nil
// 	}
//
// 	raw, err := hex.DecodeString(sigHeader.Get("v0"))
// 	if err != nil {
// 		return nil
// 	}
//
// 	return raw
// }
//
// func CHAP(w http.ResponseWriter, r *http.Request) {
// 	slackTimestamp := r.Header.Get("X-Slack-Request-Timestamp")
// 	epoch, err := strconv.Atoi(slackTimestamp)
// 	if err != nil {
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}
//
// 	reqDelta := time.Now().Sub(time.Unix(int64(epoch), 0))
// 	if reqDelta < -ChallengeTimeRange || ChallengeTimeRange < reqDelta {
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}
//
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}
//
// 	sig_basestring := fmt.Sprintf("v0:%d:%s", epoch, body)
// 	mac := hmac.New(sha256.New, []byte(Signing_Secret))
// 	mac.Write([]byte(sig_basestring))
// 	expectedMAC := mac.Sum(nil)
//
// 	if !hmac.Equal(SlackHeaderHMAC(r.Header), expectedMAC) {
// 		log.Println("[chap] Bad HMAC")
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}
//
// 	var cha struct {
// 		Challenge string
// 	}
// 	if err := json.Unmarshal(body, &cha); err != nil {
// 		log.Println("[chap]", err)
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}
//
// 	fmt.Fprintln(w, cha.Challenge)
// }
