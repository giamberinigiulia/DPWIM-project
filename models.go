package main

import (
	"time"
)

type Page struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Student struct {
	Name        string    `json:"name" sqlite:"not null"`
	Surname     string    `json:"surname" sqlite:"not null"`
	DateOfBirth time.Time `json:"date_of_birth" sqlite:"not null"`
	Username    string    `json:"username" sqlite:"primary key"`
	Password    string    `json:"password" sqlite:"not null"`
}

type Teacher struct {
	ID      int    `json:"id" sqlite:"primary key"`
	Name    string `json:"name" sqlite:"not null"`
	Surname string `json:"surname" sqlite:"not null"`
}

type Availability struct {
	ID           int       `json:"id" sqlite:"primary key"`
	Day          time.Time `json:"day" sqlite:"not null"`
	StartingTime time.Time `json:"starting_time" sqlite:"not null"`
	EndingTime   time.Time `json:"ending_time" sqlite:"not null"`
	Booked       bool      `json:"booked" sqlite:"not null"`
}

type LessonReservation struct {
	ID              int    `json:"id" sqlite:"primary key"`
	StudentUsername string `json:"student_id" sqlite:"not null"`
	TeacherID       int    `json:"teacher_id" sqlite:"not null"`
	AvailabilityID  int    `json:"availability_id" sqlite:"not null"`
	Subject         string `json:"subject" sqlite:"not null"`
}

type LessonBooked struct {
	ID             int       `json:"id" sqlite:"primary key"`
	Day            string    `json:"day" sqlite:"not null"`
	StartingTime   time.Time `json:"starting_time" sqlite:"not null"`
	EndingTime     time.Time `json:"ending_time" sqlite:"not null"`
	TeacherName    string    `json:"teacher_name" sqlite:"not null"`
	TeacherSurname string    `json:"teacher_surname" sqlite:"not null"`
	Subject        string    `json:"subject" sqlite:"not null"`
}

var errorMessage struct {
	Message string `json:"message"`
}

type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}
