package main

import "C"

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/astaxie/session/providers/memory"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	title := "Welcome to the tutoring web app"

	p := &Page{Title: title}
	t, _ := template.ParseFiles("welcome.html")
	t.Execute(w, p)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	title := "Login page"
	p := &Page{Title: title}

	t, _ := template.ParseFiles("login.html")
	t.Execute(w, p)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove the users session from the session map
	delete(sessions_new, sessionToken)

	// We need to let the client know that the cookie is expired
	// In the response, we set the session token to an empty
	// value and set its expiry as the current time
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
	// Redirect to the welcome page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	title := "Registration page"
	p := &Page{Title: title}
	ctxMsg := r.Context().Value("Message")
	if ctxMsg != nil {
		if msg, ok := ctxMsg.(string); ok {
			log.Printf("Message value: %s\n", msg)
		} else {
			log.Printf("Unexpected type for Message: %v\n", ctxMsg)
		}
	}
	t, err := template.ParseFiles("registration.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, p)
}

func userRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	dateOfBirth := r.FormValue("dateofbirth")
	password := r.FormValue("psw")
	passwordConfirm := r.FormValue("psw-repeat")
	if password != passwordConfirm {
		reloadRegistrationWithMessage(w, r, "Passwords do not match")
	} else {
		_, err := getStudentInfo(username)
		if err == nil {
			reloadRegistrationWithMessage(w, r, "Username already exists")
		} else {
			//call the Api for registration of a new student
			date, _ := time.Parse("2006-01-02", dateOfBirth)
			urlAPI := "http://localhost:8080/api/student/addstudent"
			neeStudent := Student{
				Name:        name,
				Surname:     surname,
				DateOfBirth: date,
				Username:    username,
				Password:    password,
			}
			payload, err := json.Marshal(neeStudent)
			if err != nil {
				return
			}
			resp, err := http.Post(urlAPI, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				return
			}
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)

		}
	}
}

func reloadRegistrationWithMessage(w http.ResponseWriter, r *http.Request, s string) {
	t, _ := template.New("registration.html").Funcs(timeToDate).ParseFiles("registration.html")
	t.Execute(w, &Page{Title: "Registration page", Body: s})
}

func reloadLoginWithMessage(w http.ResponseWriter, r *http.Request, s string) {
	t, _ := template.New("login.html").Funcs(timeToDate).ParseFiles("login.html")
	t.Execute(w, &Page{Title: "Login page", Body: s})
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//take r.FormValue("dateofbirth") and divide it into day, month and year and generate data
	year, err := strconv.Atoi(r.FormValue("dateofbirth")[0:4])
	if err != nil {
		return
	}
	month, err := strconv.Atoi(r.FormValue("dateofbirth")[6:7])
	if err != nil {
		return
	}
	day, err := strconv.Atoi(r.FormValue("dateofbirth")[9:10])
	if err != nil {
		return
	}
	data := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	student := &Student{
		Name:        r.FormValue("name"),
		Surname:     r.FormValue("surname"),
		DateOfBirth: data,
		Username:    r.FormValue("username"),
		Password:    r.FormValue("psw")}
	//API call at http://localhost:8080/api/student/addstudent
	payload, err := json.Marshal(student)
	if err != nil {
		return
	}

	url := "http://localhost:8080/api/student/addstudent"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userSession, err := checkSession(r)
	var student Student
	if err != nil {
		var creds Credentials
		r.ParseForm()
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")

		student, err = getStudentInfo(creds.Username)
		if err != nil {
			reloadRegistrationWithMessage(w, r, "Username doesn't found. Please register!")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(creds.Password))
		if err != nil {
			reloadLoginWithMessage(w, r, "Password doesn't match")
			return
		}

		//create a new random session token
		//we use the "github.com/google/uuid" library to generate UUIDs
		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(120 * time.Second)

		// Set the token in the session map, along with the session information
		sessions_new[sessionToken] = Session{
			username: creds.Username,
			expiry:   expiresAt,
		}

		//the client cookie for "session_token" is set using the the session token that was generated
		//the expire time is set to 120 seconds
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
		})
	} else {
		student, err = getStudentInfo(userSession.username)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	renderProfilePage(w, &student)
}

func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	userSession, err := checkSession(r)
	if err != nil {
		renderLoginPage(w, "")
	} else {
		bookings, err := getBookings(userSession.username)
		if err != nil {
			http.Error(w, "Error fetching bookings from the API", http.StatusInternalServerError)
			return
		}

		t, err := template.New("bookings.html").Funcs(timeToDate).ParseFiles("bookings.html")
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(w, struct {
			Username string
			Bookings []LessonBooked
		}{Username: userSession.username, Bookings: bookings})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func checkSession(r *http.Request) (Session, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return Session{}, errors.New("No session cookie")
		}
		return Session{}, errors.New("No session")
	}
	sessionToken := c.Value

	userSession, exists := sessions_new[sessionToken]
	if !exists {
		return Session{}, errors.New("Unauthorized: Session not found")
	}

	if userSession.isExpired() {
		delete(sessions_new, sessionToken)
		return Session{}, errors.New("Unauthorized: Session expired")
	}

	return userSession, nil
}

func deleteBookingHandler(w http.ResponseWriter, r *http.Request) {
	//retrieve ID of the booking
	id := r.FormValue("booking_id")
	urlAPI := "http://localhost:8080/api/student/bookings/" + id
	payload, err := json.Marshal(id)
	if err != nil {
		return
	}
	resp, err := http.Post(urlAPI, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var username string
	err = json.NewDecoder(resp.Body).Decode(&username)
	if err != nil {
		return
	}
	url := "http://localhost:5050/bookings?username=" + username
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func bookLessonHandler(w http.ResponseWriter, r *http.Request) {
	userSession, err := checkSession(r)
	if err != nil {
		renderLoginPage(w, "")
	} else {
		//take the list of the teachers using api
		apiURL := "http://localhost:8080/api/teachers"
		// Make a GET request to the API endpoint
		response, err := http.Get(apiURL)
		if err != nil {
			http.Error(w, "Error fetching teachers from the API", http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		//decode the JSON response into a list of teachers
		var teachers []Teacher
		err = json.NewDecoder(response.Body).Decode(&teachers)
		if err != nil {
			http.Error(w, "Error fetching teachers from the API", http.StatusInternalServerError)
			return
		}

		t, err := template.New("booklesson-teacherList.html").Funcs(timeToDate).ParseFiles("booklesson-teacherList.html")
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(w, struct {
			Username string
			Teachers []Teacher
		}{Username: userSession.username, Teachers: teachers})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func availabilityHandler(w http.ResponseWriter, r *http.Request) {
	userSession, err := checkSession(r)
	if err != nil {
		renderLoginPage(w, "")
	}
	teacherID := r.FormValue("teacher")

	//create API for retriving teacher Name and surname with the ID
	teacherName := r.FormValue("teacherName" + teacherID)
	teacherSurname := r.FormValue("teacherSurname" + teacherID)

	apiURL := "http://localhost:8080/api/teacher/" + teacherID + "/availability"
	//make a GET request to the API endpoint
	response, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Error fetching teachers from the API", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	var availabilities []Availability
	err = json.NewDecoder(response.Body).Decode(&availabilities)
	if err != nil {
		http.Error(w, "Error fetching teachers from the API", http.StatusInternalServerError)
		return
	}
	t, err := template.New("availabilityTeacher.html").Funcs(timeToDate).ParseFiles("availabilityTeacher.html")
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, struct {
		Username       string
		TeacherID      string
		TeacherName    string
		TeacherSurname string
		Availabilities []Availability
	}{Username: userSession.username, TeacherID: teacherID, TeacherName: teacherName, TeacherSurname: teacherSurname, Availabilities: availabilities})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func bookedLessonHandler(w http.ResponseWriter, r *http.Request) {
	userSession, err := checkSession(r)
	if err != nil {
		renderLoginPage(w, "")
	} else {
		r.ParseForm()
		subject := r.FormValue("subject")
		teacherID, _ := strconv.Atoi(r.Form.Get("teacherID"))
		availabilityID, _ := strconv.Atoi(r.Form.Get("selectedAvailability"))
		lesson := &LessonReservation{
			StudentUsername: userSession.username,
			TeacherID:       teacherID,
			AvailabilityID:  availabilityID,
			Subject:         subject,
		}

		//API call at http://localhost:8080/api/student/addstudent
		payload, err := json.Marshal(lesson)
		if err != nil {
			return
		}
		url := "http://localhost:8080/api/student/" + userSession.username + "/bookings"
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			http.Redirect(w, r, "/booklesson", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		}
	}
}

func getBookings(username string) ([]LessonBooked, error) {
	//construct the API endpoint URL
	apiURL := "http://localhost:8080/api/student/" + username + "/bookings"
	//make a GET request to the API endpoint
	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//decode the JSON response into the list of bookings
	var bookings []LessonBooked
	err = json.NewDecoder(response.Body).Decode(&bookings)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func renderProfilePage(w http.ResponseWriter, student *Student) {
	//render the profile page
	t, err := template.New("profile.html").Funcs(timeToDate).ParseFiles("profile.html")
	if err != nil {
		//handle template parsing error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, student)
}

func renderLoginPage(w http.ResponseWriter, errorMessage string) {
	//render the login page with an error message
	t, err := template.New("login.html").Funcs(timeToDate).ParseFiles("login.html")
	if err != nil {
		//handle template parsing error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, &Page{Title: "Login page", Body: errorMessage})
}

var timeToDate = template.FuncMap{
	"datetoFormat": func(layout string, date time.Time) string {
		return date.Format(layout)
	},
	"stringToFormat": func(date string) string {
		//transform the string to time.Time
		splittedDate := strings.Split(date, "-")
		year, _ := strconv.Atoi(splittedDate[0])
		month, _ := strconv.Atoi(splittedDate[1])
		day, _ := strconv.Atoi(strings.Split(splittedDate[2], "T")[0])
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		return t.Format("Monday, 2 January 2006")
	},
}

func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	if os.Args[1] == "-m" && len(os.Args) >= 3 {
		if os.Args[2] == "server" {
			go func() {
				defer wg.Done()
				initializeDatabase()
			}()
			gin.SetMode(gin.ReleaseMode)
			go func() {
				defer wg.Done()
				routingAPI()
			}()
		} else if os.Args[2] == "cli" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if len(os.Args) >= 4 && os.Args[3] == "-test" {
					menuCLI(true)
				} else {
					menuCLI(false)
				}
			}()
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				server()
			}()
		}
	}

	//Waiting for all the goroutines to end
	wg.Wait()
}
