package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"ttyauthz/authorization"
)

const (
	pluginName   = "ttyauthz"
	pluginFolder = "/home/danleyb2/Desktop/tmp" // "/run/docker/plugins"
)

// Manifest lists what a plugin implements.
type Manifest struct {
	// List of subsystem the plugin implements.
	Implements []string
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456"
	dbname   = "devops_tmp"
)

type User struct {
	ID       int
	Email    string
	Password string
}

func main() {
	fmt.Println("server starting..")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, dbErr := sql.Open("postgres", psqlInfo)
	if dbErr != nil {
		panic(dbErr)
	}
	defer db.Close()

	dbErr = db.Ping()
	if dbErr != nil {
		panic(dbErr)
	}
	fmt.Print("Database Connection initiated")

	// Query
	// Create an empty user and make the sql query (using $1 for the parameter)
	var myUser User
	userSql := "SELECT id, email, password FROM authentication_user WHERE id = $1"

	dbErr = db.QueryRow(userSql, 1).Scan(&myUser.ID, &myUser.Email, &myUser.Password)
	if dbErr != nil {
		log.Fatal("Failed to execute query: ", dbErr)
		//panic(dbErr)
	}
	fmt.Printf("\n User: %s, %s \n", myUser.Email, myUser.Password)

	//

	router := mux.NewRouter()

	router.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(Manifest{Implements: []string{"authz"}})

		if err != nil {
			writeErr(w, err)
			return
		}

		w.Write(b)
	})

	router.HandleFunc(fmt.Sprintf("/%s", authorization.AuthZApiRequest), func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			writeErr(w, err)
			return
		}

		var authReq authorization.Request
		err = json.Unmarshal(body, &authReq)

		if err != nil {
			writeErr(w, err)
			return
		}

		// authZRes := a.authorizer.AuthZReq(&authReq)

		var authZRes authorization.Response

		log.Fatal("Received AuthZ request, method: '%s', url: '%s'", authReq.RequestMethod, authReq.RequestURI)
		url2, err := url.Parse(authReq.RequestURI)
		if err != nil {
			authZRes = authorization.Response{
				Allow: false,
				Msg:   fmt.Sprintf("invalid request URI: %s", err.Error()),
			}
		} else {
			// action := core.ParseRoute(authReq.RequestMethod, url.Path)
			// https://github.com/twistlock/authz/blob/master/core/route_parser.go
			fmt.Println(fmt.Sprintf("Request URI: %s", url2.Path))
			action := "ActionX"

			authZRes = authorization.Response{
				Allow: false,
				Msg:   fmt.Sprintf("no policy applied (user: '%s' action: '%s')", authReq.User, action),
			}

		}

		writeResponse(w, &authZRes)
	})

	router.HandleFunc(fmt.Sprintf("/%s", authorization.AuthZApiResponse), func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			writeErr(w, err)
			return
		}

		var authReq authorization.Request
		err = json.Unmarshal(body, &authReq)

		if err != nil {
			writeErr(w, err)
			return
		}

		// authZRes := a.authorizer.AuthZRes(&authReq)
		var authZRes authorization.Response = authorization.Response{Allow: true}

		writeResponse(w, &authZRes)
	})

	// if _, err := os.Stat(pluginFolder); os.IsNotExist(err) {
	//   fmt.Println("Creating plugins folder %q", pluginName)
	//   err = os.MkdirAll("/run/docker/plugins/", 0750)
	//   if err != nil {
	//     log.Fatal( err)
	//   }
	// }

	pluginPath := fmt.Sprintf("%s/%s.sock", pluginFolder, pluginName)
	os.Remove(pluginPath)

	var listener net.Listener
	var err error

	listener, err = net.ListenUnix("unix", &net.UnixAddr{Name: pluginPath, Net: "unix"})
	if err != nil {
		// return err
		log.Fatal(err)
	} else {
		defer listener.Close()
		http.Serve(listener, router)
	}

}

// writeResponse writes the authZPlugin response to response writer
func writeResponse(w http.ResponseWriter, authZRes *authorization.Response) {

	data, err := json.Marshal(authZRes)
	if err != nil {
		log.Fatal("Failed to marshel authz response %q", err.Error())
	} else {
		w.Write(data)
	}

	if authZRes == nil || authZRes.Err != "" {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// writeErr writes the authZPlugin error response to response writer
func writeErr(w http.ResponseWriter, err error) {
	writeResponse(w, &authorization.Response{Err: err.Error()})
}
