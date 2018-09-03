package mcstats

import (
	"fmt"
	"strings"
	"time"

	"github.com/renevo/mcstats/pkg/simplehttp"
)

// GetName will lookup the mojang account by uid for the specified timestamp
//
// Names can change over time, so we want the correct one
func GetName(uid string, ts time.Time) (string, error) {
	nameHistory := []struct {
		Name    string `json:"name"`
		Changed int64  `json:"changedToAt"`
	}{}

	uidString := strings.ToLower(strings.Replace(uid, "-", "", -1))
	if err := simplehttp.GetJSON(fmt.Sprintf("https://api.mojang.com/user/profiles/%s/names", uidString), &nameHistory); err != nil {
		return "", fmt.Errorf("failed to get profile %s from mojang API: %v", uidString, err)
	}

	cts := ts.Unix()
	current := ""

	// changed of zero means it was the first one the account was created with
	// past that, each item in the array will have a unix epoch of when it was changed
	// we want to start with the first one, and capture the one that is changed >= <

	for _, hist := range nameHistory {
		if hist.Changed <= cts {
			current = hist.Name
		}
	}

	return current, nil
}
