func receiveData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "ERROR", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "ERROR", http.StatusBadRequest)
		return
	}

	messageRaw, ok := data["message"]
	if !ok {
		response := Response{
			Status:  "400",
			Message: "ERROR",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	message, ok := messageRaw.(string)
	if !ok {
		response := Response{
			Status:  "400",
			Message: "ERROR",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	fmt.Println("MESSAGE:", message)

	response := Response{
		Status:  "success",
		Message: "SUCCESS",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
