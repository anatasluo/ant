package setting

import "net/http"

func HandleDeny(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "Permission denied!", http.StatusForbidden)
}

