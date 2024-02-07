package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func routingAPI() {
	fmt.Println("API server is running on port 8080")

	router := gin.Default() // Using gin.Default() to set up the default middleware

	api := router.Group("/api")

	teachersGroup := api.Group("/teachers")
	teachersGroup.GET("", getTeachers)
	teachersGroup.GET("/:name/:surname", getTeacherIDByNameAndSurname)
	teachersGroup.POST("/addteacher", createNewTeacher)

	teacherGroup := api.Group("/teacher")
	teacherGroup.GET("/:id/availability", getTeacherAvailability)
	teacherGroup.GET("/:id/bookings", getTeacherBookings)
	teacherGroup.POST("/:id/availability", createTeacherAvailability)

	studentGroup := api.Group("/student")
	studentGroup.POST("/addstudent", createNewStudent)
	studentGroup.GET("/allstudents", getStudents)
	studentGroup.GET("/:username/profile", getProfileStudent)
	studentGroup.GET("/:username/bookings", getStudentBookings)
	studentGroup.POST("/:username/bookings", createStudentBooking)
	studentGroup.POST("/bookings/:id", deleteStudentBooking)

	// Run the server on port 8080
	router.Run("localhost:8080")
}

func server() {
	fmt.Println("Web server is running on port 5050")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/registration", registrationHandler)
	http.HandleFunc("/userregistration", userRegistrationHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/bookings", bookingsHandler)
	http.HandleFunc("/deleteBooking", deleteBookingHandler)
	http.HandleFunc("/booklesson", bookLessonHandler)
	http.HandleFunc("/availability", availabilityHandler)
	http.HandleFunc("/bookedLesson", bookedLessonHandler)

	// Run the server on port 5050
	http.ListenAndServe("localhost:5050", nil)
}
