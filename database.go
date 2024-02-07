package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// Database setup and connection
func initializeDatabase() {
	fmt.Println("Database connection...")
	connectToDB()
	defer db.Close()

	createTables()
}

func connectToDB() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS students (
			Name TEXT NOT NULL,
			Surname TEXT NOT NULL,
			DateOfBirth DATE NOT NULL,
			Username TEXT NOT NULL UNIQUE,
			Password TEXT NOT NULL,
			PRIMARY KEY (Username)
		)`,
		`CREATE TABLE IF NOT EXISTS teachers (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Surname TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS availabilities (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			TeacherID INTEGER NOT NULL,
			Day DATE NOT NULL,
			StartingTime DATE NOT NULL,
			EndingTime DATE NOT NULL,
			Booked BOOLEAN NOT NULL,
			FOREIGN KEY (TeacherID) REFERENCES teachers(ID)
		)`,
		`CREATE TABLE IF NOT EXISTS bookings (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			StudentUsername TEXT NOT NULL,
			TeacherID INTEGER NOT NULL,
			AvailabilityID INTEGER NOT NULL,
			Subject TEXT NOT NULL,
			FOREIGN KEY (StudentUsername) REFERENCES students(Username),
			FOREIGN KEY (TeacherID) REFERENCES teachers(ID),
			FOREIGN KEY (AvailabilityID) REFERENCES availabilities(ID)
		)`,
	}

	for _, table := range tables {
		err := createTableIfNotExists(db, table)
		if err != nil {
			printMessage("Error creating table: " + err.Error())
			return
		}
	}
}

func createTableIfNotExists(db *sql.DB, tableDefinition string) error {
	_, err := db.Exec(fmt.Sprintf(`
		%s
	`, tableDefinition))
	return err
}

// Getters methods
// getAvailabilityByID returns the availability
func getAvailabilityByID(db *sql.DB, id int) (Availability, error) {
	var availability Availability
	row := db.QueryRow("SELECT ID, Day, StartingTime, EndingTime, Booked FROM availabilities WHERE ID =?", id)
	err := row.Scan(&availability.ID, &availability.Day, &availability.StartingTime, &availability.EndingTime, &availability.Booked)
	return availability, err
}

// getTeacherIDByFullName retrieves the ID of a teacher by their full name from the database.
func getTeacherIDByFullName(db *sql.DB, name, surname string) (int, error) {
	var teacherID int

	row := db.QueryRow(`
        SELECT ID FROM teachers
        WHERE Name = ? AND Surname = ?
    `, name, surname)

	err := row.Scan(&teacherID)
	if err != nil {
		return 0, err
	}

	return teacherID, nil
}

// getTeacherAvailabilities retrieves the availabilities of a teacher from the database.
func getTeacherAvailabilities(db *sql.DB, teacherID int) ([]Availability, error) {
	isPresent, err := isTeacherExists(db, teacherID)
	if err != nil {
		log.Fatal(err)
	}
	if !isPresent {
		return nil, &ErrTeacherNotFound{TeacherID: teacherID}
	}
	var availabilities []Availability

	rows, err := db.Query(`
		SELECT ID, Day, StartingTime, EndingTime, Booked
		FROM availabilities
		WHERE TeacherID = ?
	`, teacherID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var availability Availability
		err := rows.Scan(&availability.ID, &availability.Day, &availability.StartingTime, &availability.EndingTime, &availability.Booked)
		if err != nil {
			return nil, err
		}
		availabilities = append(availabilities, availability)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return availabilities, nil
}

// getAllTeachers retrieves all teachers from the database.
func getAllTeachers(db *sql.DB) ([]Teacher, error) {
	var teachers []Teacher

	rows, err := db.Query("SELECT ID, Name, Surname FROM teachers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var teacher Teacher
		err := rows.Scan(&teacher.ID, &teacher.Name, &teacher.Surname)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

// getTeacherAvailabilitiesByID retrieves the bookings of a teacher by their ID from the database.
func getTeacherAvailabilitiesByID(db *sql.DB, teacherID int) ([]Availability, error) {
	isPresent, err := isTeacherExists(db, teacherID)
	if err != nil {
		return nil, err
	}
	if !isPresent {
		return nil, &ErrTeacherNotFound{TeacherID: teacherID}
	}
	var availabilities []Availability

	rows, err := db.Query(`
        SELECT ID, Day, StartingTime, EndingTime, Booked
        FROM availabilities
        WHERE TeacherID = ? AND Booked = 1
    `, teacherID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var availability Availability
		err := rows.Scan(&availability.ID, &availability.Day, &availability.StartingTime, &availability.EndingTime, &availability.Booked)
		if err != nil {
			return nil, err
		}
		availabilities = append(availabilities, availability)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return availabilities, nil
}

// getAllStudents retrieves all students from the database.
func getAllStudents(db *sql.DB) ([]Student, error) {
	var students []Student

	rows, err := db.Query("SELECT * FROM students")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var student Student
		err := rows.Scan(&student.Name, &student.Surname, &student.DateOfBirth, &student.Username, &student.Password)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

// getStudentByUsername retrieves a student by their username from the database.
func getStudentByUsername(db *sql.DB, username string) (Student, error) {
	var student Student
	var date time.Time

	row := db.QueryRow("SELECT * FROM students WHERE Username =?", username)
	err := row.Scan(&student.Name, &student.Surname, &date, &student.Username, &student.Password)

	if err == sql.ErrNoRows {
		// No student found with the specified username
		return Student{}, &ErrStudentNotFound{StudentID: username}
	} else if err != nil {
		// Handle other errors
		return Student{}, err
	}

	dateOfBirthFormatted := date.Format("02/01/2006")
	student.DateOfBirth, err = time.Parse("02/01/2006", dateOfBirthFormatted)
	if err != nil {
		return Student{}, err
	}

	return student, nil
}

// deleteBookingByID deletes a booking by its ID from the database.
func deleteBookingByID(db *sql.DB, id string) (string, error) {
	var availabilityID int
	var studentUsername string

	row := db.QueryRow(`
        SELECT AvailabilityID, StudentUsername
        FROM bookings
        WHERE ID =?
    `, id)
	err := row.Scan(&availabilityID, &studentUsername)
	if err != nil {
		return "", err
	}

	// Update availability status booked to false
	stmt, err := db.Prepare("UPDATE availabilities SET Booked = 0 WHERE ID =?")
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(availabilityID)
	if err != nil {
		return "", err
	}

	// Delete the booking
	stmt, err = db.Prepare("DELETE FROM bookings WHERE id =?")
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return "", err
	}

	return studentUsername, nil
}

// getStudentBookingsByUsername retrieves the bookings of a student by their username from the database.
func getStudentBookingsByUsername(db *sql.DB, studentUsername string) ([]LessonBooked, error) {
	// Check if the student exists
	isPresent, err := isStudentExists(db, studentUsername)
	if err != nil {
		return nil, err
	}
	if !isPresent {
		return nil, &ErrStudentNotFound{StudentID: studentUsername}
	}

	query := `
        SELECT
            b.ID AS id,
            a.Day AS day,
            a.StartingTime AS starting_time,
            a.EndingTime AS ending_time,
            t.Name AS teacher_name,
            t.Surname AS teacher_surname,
            b.Subject AS subject
        FROM
            bookings b
        JOIN
            availabilities a ON b.AvailabilityID = a.ID
        JOIN
            teachers t ON b.TeacherID = t.ID
        JOIN
            students u ON b.StudentUsername = u.Username
        WHERE
            u.Username = ?;
    `

	rows, err := db.Query(query, studentUsername)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var bookings []LessonBooked

	for rows.Next() {
		var booking LessonBooked
		// Scan and parse the data
		err := rows.Scan(&booking.ID, &booking.Day, &booking.StartingTime, &booking.EndingTime, &booking.TeacherName, &booking.TeacherSurname, &booking.Subject)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

// Insert methods

// insertTeacher inserts a new teacher into the database.
func insertTeacher(db *sql.DB, teacher Teacher) error {
	_, err := db.Exec(`
		INSERT INTO teachers (Name, Surname)
		VALUES (?, ?)
	`, teacher.Name, teacher.Surname)
	return err
}

// insertAvailability inserts a new availability for a teacher into the database.
func insertAvailability(db *sql.DB, availability Availability, teacherID int) error {
	// Check if the teacher exists
	isPresent, err := isTeacherExists(db, teacherID)
	if err != nil {
		return err
	}

	if !isPresent {
		return &ErrTeacherNotFound{TeacherID: teacherID}
	}

	// Check for overlapping availabilities
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM availabilities
		WHERE TeacherID = ? AND Day = ? AND (
			(StartingTime <= ? AND EndingTime > ?) OR
			(StartingTime < ? AND EndingTime >= ?) OR
			(StartingTime >= ? AND EndingTime <= ?)
		)
	`, teacherID, availability.Day, availability.StartingTime, availability.StartingTime, availability.EndingTime, availability.EndingTime, availability.StartingTime, availability.EndingTime).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		// Overlapping availabilities
		return errors.New("Overlapping availabilities")
	}

	_, err = db.Exec(`
		INSERT INTO availabilities (TeacherID, Day, StartingTime, EndingTime, Booked)
		VALUES (?, ?, ?, ?, ?)
	`, teacherID, availability.Day, availability.StartingTime, availability.EndingTime, availability.Booked)

	return err
}

// insertStudent inserts a new student into the database.
func insertStudent(db *sql.DB, student Student) error {
	// Hash the password
	hashedPassword, err := hashPassword(student.Password)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO students (Name, Surname, DateOfBirth, Username, Password)
        VALUES (?,?,?,?,?)
    `, student.Name, student.Surname, student.DateOfBirth.Format("2006-01-02"), student.Username, hashedPassword)

	if err != nil {
		// Check if the error is due to a unique constraint violation
		if sqliteErr, ok := err.(*sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return errors.New("Username already exists")
		}
		return err
	}

	return nil
}

// insertBooking inserts a new booking into the database.
func insertBooking(db *sql.DB, booking LessonReservation) error {
	// Check if the student and the teacher of the booking exists
	isPresent, err := isStudentExists(db, booking.StudentUsername)
	if err != nil {
		return err
	}
	if !isPresent {
		return &ErrStudentNotFound{StudentID: booking.StudentUsername}
	}

	isPresent, err = isTeacherExists(db, booking.TeacherID)
	if err != nil {
		return err
	}
	if !isPresent {
		return &ErrTeacherNotFound{TeacherID: booking.TeacherID}
	}

	isPresent, err = isAvailabilityRelatedToTeacher(db, booking.AvailabilityID, booking.TeacherID)
	if err != nil {
		return err
	}
	if !isPresent {
		return errors.New("Availability not related to the teacher")
	}

	// Check if the availability is already booked
	var count int
	err = db.QueryRow(`
        SELECT COUNT(*)
        FROM availabilities
        WHERE ID =? AND Booked = 1
    `, booking.AvailabilityID).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("Availability already booked")
	}

	availability, err := getAvailabilityByID(db, booking.AvailabilityID)
	if err != nil {
		return err
	}
	// Check for overlapping times with other bookings made by the same student
	var overlappingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) AS OverlappingCount
		FROM bookings b
		JOIN availabilities a ON b.AvailabilityID = a.ID
		WHERE b.StudentUsername = ? AND b.AvailabilityID <> ? AND Day = ? AND (
			(a.StartingTime <= ? AND a.EndingTime > ?) OR
			(a.StartingTime < ? AND a.EndingTime >= ?) OR
			(a.StartingTime >= ? AND a.EndingTime <= ?)
		)
		`, booking.StudentUsername, booking.AvailabilityID, availability.Day,
		availability.StartingTime, availability.EndingTime,
		availability.StartingTime, availability.EndingTime,
		availability.StartingTime, availability.EndingTime).Scan(&overlappingCount)

	if err != nil {
		return err
	}

	if overlappingCount > 0 {
		return errors.New("Overlapped times with existing bookings for the same student")
	}

	_, err = db.Exec(`
        INSERT INTO bookings (StudentUsername, TeacherID, AvailabilityID, Subject)
        VALUES (?,?,?,?)
    `, booking.StudentUsername, booking.TeacherID, booking.AvailabilityID, booking.Subject)

	if err != nil {
		return err
	}

	// Update booking availability status
	_, err = db.Exec(`
        UPDATE availabilities
        SET Booked = 1
        WHERE ID =?
    `, booking.AvailabilityID)

	if err != nil {
		return err
	}

	return err
}

// Utilities methods

// isTeacherExists checks if a teacher with the given ID exists in the database.
func isTeacherExists(db *sql.DB, teacherID int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM teachers WHERE ID = ?)", teacherID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// isStudentExists checks if a student with the given username exists in the database.
func isStudentExists(db *sql.DB, studentUsername string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM students WHERE Username = ?)", studentUsername).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// isAvailabilityRelatedToTeacher checks if an availability with the given ID is related to the specified teacher.
func isAvailabilityRelatedToTeacher(db *sql.DB, availabilityID int, teacherID int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM availabilities WHERE ID =? AND TeacherID =?)", availabilityID, teacherID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// hashPassword hashes the given password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
