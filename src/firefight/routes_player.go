package firefight

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Join(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	scmd, ok := r.Context().Value("slack_cmd").(*SlackCmd)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if err := ff.Join(scmd.UserID); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "ephemeral",
			Text: "You've joined the fight!",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[Join]", err)
	}
}

func Target(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	scmd, ok := r.Context().Value("slack_cmd").(*SlackCmd)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if target, err := ff.GetTarget(scmd.UserID); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "ephemeral",
			Text: fmt.Sprintf("Your next target: <@%s>.", target.ID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[Target]", err)
	}
}

func ReportHit(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	scmd, ok := r.Context().Value("slack_cmd").(*SlackCmd)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if target, err := ff.ReportHit(scmd.UserID); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "in_channel",
			Text: fmt.Sprintf("<@%s> has been hit!", target.ID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[ReportHit]", err)
	}
}

func DisputeHit(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	scmd, ok := r.Context().Value("slack_cmd").(*SlackCmd)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if err := ff.DisputeHit(scmd.UserID); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "in_channel",
			Text: fmt.Sprintf("FFbot revived: <@%s>.", scmd.UserID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[DisputeHit]", err)
	}
}

func DefendAttack(w http.ResponseWriter, r *http.Request) {
	ff, ok := r.Context().Value("fire_fight").(*FireFight)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	scmd, ok := r.Context().Value("slack_cmd").(*SlackCmd)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data SlackResponse
	if _, err := ff.Defend(scmd.UserID); err != nil {
		data = SlackResponse{Type: "ephemeral", Text: err.Error()}
	} else {
		data = SlackResponse{
			Type: "in_channel",

			// Likely don't want to reveal if the defence was correct?
			Text: fmt.Sprintf("<@%s> defended an attack.", scmd.UserID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("[DefendAttack]", err)
	}
}
