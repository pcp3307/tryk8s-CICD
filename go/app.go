/*
#####################################
# 2018 ithome ironman
# Author: James Lu
# Topic: k8s 不自賞 - Day 26 部署多階層應用：簡易通訊錄 SAB
# Url: https://ithelp.ithome.com.tw/articles/10195946
# Licence: MIT
#####################################
*/
package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
 * Structure for Person
 */
type Person struct {
	Name  string
	Email string
}

/*
 * Structure for Home Page Info
 */
type HomeInfo struct {
	IP     string
	Member []Person
}

/*
 * Structure for Error Info
 */
type ErrorInfo struct {
	IP  string
	Msg string
}

const dbName = "info"
const dbCollection = "people"
const defaultServerPort = "80"
const defaultMongoDb = "127.0.0.1:27017"

/*
 * Landing Page
 * 1. load /pages/home.html as template
 * 2. load container ip
 * 3. load records in db
 */
func landing(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("./pages/home.html")
	if err != nil {
		log.Fatal(err)
	}

	info := HomeInfo{loadIP(), loadInfo()}

	err = t.Execute(w, info)
	if err != nil {
		log.Fatal(err)
	}
}

/*
 * Load Container IP Address
 * 1. load ip
 * 2. construct ip according to /etc/hosts if ip does not exist
 */
func loadIP() string {
	ip, errLoadFile := ioutil.ReadFile("ip")
	if errLoadFile != nil {
		out, err := exec.Command("sh", "-c", "awk 'END{print $1}' /etc/hosts").Output()
		if err != nil {
			panic(err)
		}

		//write to ip file
		err = ioutil.WriteFile("./ip", out, 0644)
		if err != nil {
			panic(err)
		}

		return string(out)
	}

	return string(ip)
}

/*
 * Get Mongodb Server Url
 */
func getMongoDb() string {
	db := os.Getenv("DB_SERVER")
	if db == "" {
		db = defaultMongoDb
	}
	return db
}

/*
 * Load Records from DB
 * 1. connect Mongodb according env DB_SERVER
 * 2. load records from 'people' collection of 'info' db in Mongodb
 */
func loadInfo() []Person {
	session, err := mgo.Dial(getMongoDb())

	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB(dbName).C(dbCollection)

	var result []Person

	err = c.Find(nil).All(&result)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

/*
 * Save Page
 * 1. process POST request
 * 2. save record to mongodb
 */
func saveHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if "POST" == r.Method {
		errSave := saveToDB(Person{r.FormValue("name"), r.FormValue("email")})
		if errSave != nil {
			t, errLoadTmp := template.ParseFiles("./pages/err.html")
			if errLoadTmp != nil {
				log.Fatal(errLoadTmp)
			}

			errExec := t.Execute(w, ErrorInfo{loadIP(), errSave.Error()})

			if errExec != nil {
				log.Fatal(errExec)
			}

			return
		}

	}

	http.Redirect(w, r, "/", http.StatusFound)
}

/*
 * Save record to mongodb
 * 1. check record exist
 * 2. save record to mongodb
 */
func saveToDB(p Person) error {

	session, err := mgo.Dial(getMongoDb())

	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB(dbName).C(dbCollection)

	user := Person{}
	query := c.Find(bson.M{"name": p.Name, "email": p.Email}).One(&user)
	if query != nil {
		errInsert := c.Insert(&p)

		if errInsert != nil {
			log.Fatal(errInsert)
		}

		return nil
	}

	return errors.New(p.Name + "：" + p.Email + " 已存在")
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = defaultServerPort
	}

	http.HandleFunc("/", landing)
	http.HandleFunc("/save", saveHandler)
	http.ListenAndServe(":"+port, nil)
}
