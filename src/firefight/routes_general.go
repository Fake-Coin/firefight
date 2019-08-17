package firefight

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type SlackResponse struct {
	Type string `json:"response_type,omitempty"`
	Text string `json:"text,omitempty"`
}

func Start(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if err := ff.Start(); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "in_channel",
			Text: "FireFight Started!",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[Start]", err)
	}
}

func Pause(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if err := ff.Pause(); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "in_channel",
			Text: "[Paused] Ceasefire!",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[Pause]", err)
	}
}

func End(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if players, err := ff.End(); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		var finalScores strings.Builder
		finalScores.WriteString("[FireFight Scoreboard]\n")
		for i, p := range players {
			finalScores.WriteString(fmt.Sprintf("#%d: % 2dpts - <@%s>\n", i, p.Score, p.ID))
		}

		data = SlackResponse{Type: "in_channel", Text: finalScores.String()}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[End]", err)
	}
}

func Scoreboard(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var topPlayers strings.Builder
	topPlayers.WriteString("[FireFight Scoreboard]\n")

	players := ff.Scoreboard()
	for i, p := range players {
		status := "active"
		if p.Hit {
			status = "fragged"
		}

		topPlayers.WriteString(fmt.Sprintf("#%d: % 2dpts - <@%s> (%s)\n", i, p.Score, p.ID, status))
	}

	data := SlackResponse{Type: "ephemeral", Text: topPlayers.String()}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[Scoreboard]", err)
	}
}
