package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// createNewStudent creates a new student using the provided JSON data.
func createNewStudent(c *gin.Context) {
	connectToDB()

	var newStudent Student

	//decode JSON request body to create a new student
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&newStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing JSON"})
		return
	}

	//insert the new student into the database
	err := insertStudent(db, newStudent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Student created successfully"})
}

// getStudents retrieves a list of all students.
func getStudents(c *gin.Context) {
	connectToDB()
	students, err := getAllStudents(db)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No students")})
		return
	}

	if len(students) == 0 {
		c.JSON(http.StatusOK, []Student{})
		return
	}

	c.IndentedJSON(http.StatusOK, students)
}

// getProfileStudent retrieves the profile of a specific student using their username.
func getProfileStudent(c *gin.Context) {
	connectToDB()

	//retrieve the username of the student from the URL parameter
	username := c.Param("username")

	student, err := getStudentByUsername(db, username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Student not found")})
		return
	}

	c.IndentedJSON(http.StatusOK, student)
}

// createStudentBooking creates a new booking for a student.
func createStudentBooking(c *gin.Context) {
	connectToDB()

	var newBooking LessonReservation

	//decode JSON request body to create a new booking
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&newBooking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing JSON"})
		return
	}

	//insert the new booking into the database
	err := insertBooking(db, newBooking)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Booking created successfully"})
}

// getStudentBookings retrieves all bookings for a specific student using their username.
func getStudentBookings(c *gin.Context) {
	connectToDB()

	//retrieve the username of the student from the URL parameter
	username := c.Param("username")

	bookings, err := getStudentBookingsByUsername(db, username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Student not found")})
		return
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusOK, []LessonBooked{})
		return
	}

	c.IndentedJSON(http.StatusOK, bookings)
}

// deleteStudentBooking deletes a booking for a student using the booking ID.
func deleteStudentBooking(c *gin.Context) {
	connectToDB()

	//retrieve the ID for the lessonBooked from the URL parameter
	id := c.Param("id")

	//delete the booking and retrieve the student's username
	username, err := deleteBookingByID(db, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error deleting booking"})
		return
	}

	c.IndentedJSON(http.StatusOK, username)
}
