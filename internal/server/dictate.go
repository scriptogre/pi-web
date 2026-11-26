package server

import (
	"cmp"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// handleDictate mints a 30-second Deepgram token, so the browser streams audio
// to Deepgram directly and never sees the key. It reads the same key and
// language as the pi-dictate extension, so terminal and web stay in sync.
func (s *Server) handleDictate(w http.ResponseWriter, r *http.Request) {
	var auth struct{ Deepgram struct{ Key string } }
	var dictate struct{ Language string }
	readAgentJSON(s.agentDir, "auth.json", &auth)
	readAgentJSON(s.agentDir, "pi-dictate.json", &dictate)
	if auth.Deepgram.Key == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "no Deepgram key in auth.json")
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://api.deepgram.com/v1/auth/grant", nil)
	req.Header.Set("Authorization", "Token "+auth.Deepgram.Key)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer res.Body.Close()
	var grant struct {
		AccessToken string `json:"access_token"`
	}
	if json.NewDecoder(res.Body).Decode(&grant); grant.AccessToken == "" {
		writeJSONError(w, http.StatusBadGateway, "Deepgram token grant failed: "+res.Status)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": grant.AccessToken, "language": cmp.Or(dictate.Language, "en")})
}

// readAgentJSON leaves out unchanged when the file is missing or invalid.
func readAgentJSON(agentDir, name string, out any) {
	if data, err := os.ReadFile(filepath.Join(agentDir, name)); err == nil {
		_ = json.Unmarshal(data, out)
	}
}
