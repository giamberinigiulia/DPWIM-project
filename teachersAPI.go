package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Getters

// getTeachers retrieves a list of all teachers.
func getTeachers(c *gin.Context) {
	connectToDB()
	teachers, err := getAllTeachers(db)
	if err != nil {
		log.Fatal(err)
	}
	if len(teachers) == 0 {
		c.JSON(http.StatusOK, []Teacher{})
		return
	}
	c.IndentedJSON(http.StatusOK, teachers)
}

// getTeacherAvailability retrieves the availabilities of a specific teacher using their ID.
func getTeacherAvailability(c *gin.Context) {
	connectToDB()
	teacherID, errID := strconv.Atoi(c.Param("id"))
	if errID != nil {
		log.Fatal(errID)
	}
	isPresent, _ := isTeacherExists(db, teacherID)
	if !isPresent {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No teachers associated with ID %d", teacherID)})
		return
	}
	availabilities, err := getTeacherAvailabilities(db, teacherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No availabilities")})
		return
	}
	if len(availabilities) == 0 {
		c.JSON(http.StatusOK, []Availability{})
		return
	}
	c.IndentedJSON(http.StatusOK, availabilities)
}

// getTeacherBookings retrieves the bookings of a specific teacher using their ID.
func getTeacherBookings(c *gin.Context) {
	connectToDB()
	teacherID, errID := strconv.Atoi(c.Param("id"))
	if errID != nil {
		log.Fatal(errID)
	}
	//find out if the teacher is saved in the DB
	isPresent, _ := isTeacherExists(db, teacherID)
	if !isPresent {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No teachers associated with ID %d", teacherID)})
		return
	}
	//retrieve all the availabilities of the teacher
	bookings, err := getTeacherAvailabilitiesByID(db, teacherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No bookings")})
		return
	}
	if len(bookings) == 0 {
		c.JSON(http.StatusOK, []Availability{})
		return
	}
	c.IndentedJSON(http.StatusOK, bookings)
}

// getTeacherIDByNameAndSurname retrieves the ID of a teacher using their name and surname.
func getTeacherIDByNameAndSurname(c *gin.Context) {
	connectToDB()
	teacherName := c.Param("name")
	teacherSurname := c.Param("surname")
	teacherID, errID := getTeacherIDByFullName(db, teacherName, teacherSurname)
	if errID != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No teachers associated with name %s and surname %s", teacherName, teacherSurname)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": teacherID})
}

// Creators

// createNewTeacher creates a new teacher using the provided JSON data.
func createNewTeacher(c *gin.Context) {
	connectToDB()
	var newTeacher Teacher

	if err := c.ShouldBindJSON(&newTeacher); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR"})
		return
	}
	insertTeacher(db, newTeacher)

	c.JSON(http.StatusCreated, gin.H{"message": "Teacher created successfully"})
}

// createTeacherAvailability creates a new availability for a teacher.
func createTeacherAvailability(c *gin.Context) {
	connectToDB()
	defer db.Close()

	teacherID, errID := strconv.Atoi(c.Param("id"))
	if errID != nil {
		log.Fatal(errID)
	}
	//find out if the teacher is saved in the DB
	isPresent, _ := isTeacherExists(db, teacherID)
	if !isPresent {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("No teachers associated with ID %d", teacherID)})
		return
	}

	var availability Availability

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&availability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing JSON"})
		return
	}
	//checking if the duration of the lesson is 1 hour
	checkDuration, _ := checkDuration(availability.StartingTime, availability.EndingTime)
	if !checkDuration {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("The duration of each lesson need to be 1 hour")})
		return
	}

	err := insertAvailability(db, availability, teacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error creating new availability"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Availability created successfully"})
}

// Utils

// checkDuration checks if the duration between starting and ending times is exactly 1 hour.
func checkDuration(startingTime, endingTime time.Time) (bool, error) {
	duration := endingTime.Sub(startingTime)
	if duration.Hours() != 1 {
		return false, fmt.Errorf("The duration of the lesson need to be of 1 hour")
	}
	return true, nil
}
