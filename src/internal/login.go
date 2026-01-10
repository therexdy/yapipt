package internal

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"yapipt/pkg"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)


func (R *Runtime)Login(w http.ResponseWriter, r *http.Request){
	var req_json pkg.LoginJSON
	err := json.NewDecoder(r.Body).Decode(&req_json)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		pkg.LogInfo("Error during json parsing " + err.Error())
		return
	}

	_, err = R.RedisDB.Get(R.DBContext, req_json.UserName).Result()
	if err != nil{
		if err != redis.Nil {
			w.WriteHeader(http.StatusInternalServerError)
			pkg.LogInfo("Error during Redis Query" + err.Error())
			return
		}
		R.RedisDB.Del(R.DBContext, req_json.UserName)
		R.HubMutex.Lock()
		delete(R.WSConnHub, req_json.UserName)
		R.HubMutex.Unlock()
	}

	var resp_json pkg.SessionTokenJSON

	var password_from_db string
	err = R.PSQL_DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", req_json.UserName).Scan(&password_from_db)
	if err == sql.ErrNoRows {
		hashed_password, err := HashPassword(req_json.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			pkg.LogInfo("Error during Hashing" + err.Error())
			return
		}
		_, err = R.PSQL_DB.Exec("INSERT INTO users(username, password_hash) VALUES ($1, $2)", req_json.UserName, hashed_password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			pkg.LogInfo("Error during INSERT "+err.Error())
			return
		}
		session_token, err := R.NewSessionToken(req_json.UserName)
		if err != nil {
			pkg.LogError("Could not generate session_token "+err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp_json = pkg.SessionTokenJSON{UserName: req_json.UserName, Success: true, SessionToken: session_token}
		json.NewEncoder(w).Encode(resp_json)
		return
	} else if err != nil {
		pkg.LogError("Query error for password "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {

		match, err := VerifyPassword(req_json.Password, password_from_db)

		if !match {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		session_token, err := R.NewSessionToken(req_json.UserName)
		if err != nil {
			pkg.LogError("Could not generate session_token" + err.Error())
			return
		}
		resp_json = pkg.SessionTokenJSON{UserName: req_json.UserName, Success: true, SessionToken: session_token}
		json.NewEncoder(w).Encode(resp_json)
		return
	}
}

