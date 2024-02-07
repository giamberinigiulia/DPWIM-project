package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func menuCLI(test bool) {
	fmt.Println("Welcome to the Menu!")

	for {
		printMenu(test)
		var message string
		if test {
			message = "Select an option (0-9): "
		} else {
			message = "Select an option (0-4): "
		}
		option := getUserInput(message)

		switch option {
		case "1":
			fmt.Println("Adding a teacher...")
			var teacher Teacher
			//retrieve data from cli for creating a teacher
			teacher.Name = getUserInput("Enter the teacher's name: ")
			teacher.Surname = getUserInput("Enter the teacher's surname: ")

			//api call
			url := "http://localhost:8080/api/teachers/addteacher"
			payload, err := json.Marshal(teacher)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}

			if resp.StatusCode != http.StatusCreated {
				err := json.Unmarshal([]byte(body), &errorMessage)
				if err != nil {
					printErrorMessage(err, "Error parsing JSON: ")
					return
				}
				printMessage("Some error occurred: " + errorMessage.Message)
				break
			} else {
				printMessage("Teacher added successfully!")
			}

		case "2":
			fmt.Println("Adding an availability for a specific teacher...")
			var teacher Teacher
			//retrieve data from cli for creating an availability
			teacher.Name = getUserInput("Enter the teacher's name: ")
			teacher.Surname = getUserInput("Enter the teacher's surname: ")

			//api call
			baseUrl := "http://localhost:8080/api/teachers/" + teacher.Name + "/" + teacher.Surname + "/"
			resp, err := http.Get(baseUrl)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				printMessage("No teacher found as " + teacher.Name + " " + teacher.Surname)
				break
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}

			err = json.Unmarshal(body, &teacher)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}

			//retrieve day, endingTime and starting time from the cli
			date := getUserInput("Enter the day: ")
			dayStr := date[0:2]
			monthStr := date[3:5]
			yearStr := date[6:10]

			//convert string components to integers
			day, err := strconv.Atoi(dayStr)
			if err != nil {
				printErrorMessage(err, "Error converting day: ")
				break
			}

			month, err := strconv.Atoi(monthStr)
			if err != nil {
				printErrorMessage(err, "Error converting month: ")
				break
			}

			year, err := strconv.Atoi(yearStr)
			if err != nil {
				printErrorMessage(err, "Error converting year: ")
				break
			}
			if !isValidDate(year, month, day) {
				printMessage("Invalid date")
				break
			}

			//create a time.Time value using the components
			parsedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

			startingTime := getUserInput("Enter the starting time: ")
			hour, err := strconv.Atoi(startingTime[0:2])
			if err != nil {
				printErrorMessage(err, "Error converting hour: ")
				break
			}
			min, err := strconv.Atoi(startingTime[3:5])
			if err != nil {
				printErrorMessage(err, "Error converting minute: ")
				break
			}
			parsedStartingTIme := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)

			endingTime := getUserInput("Enter the starting time: ")
			hour, err = strconv.Atoi(endingTime[0:2])
			if err != nil {
				printErrorMessage(err, "Error converting hour: ")
				break
			}
			min, err = strconv.Atoi(endingTime[3:5])
			if err != nil {
				printErrorMessage(err, "Error converting minute: ")
				break
			}
			parsedEndingTime := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)

			//generate the availability with booked false
			availability := Availability{Day: parsedDate, StartingTime: parsedStartingTIme, EndingTime: parsedEndingTime, Booked: false}

			//insert it into the database using a POST request
			baseUrl = fmt.Sprintf("http://localhost:8080/api/teacher/%d/availability", teacher.ID)
			payload, err := json.Marshal(availability)
			if err != nil {
				printMessage("It wasn't possible to encode the availability to JSON")
				break
			}
			resp, err = http.Post(baseUrl, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				printErrorMessage(err, "The availability couldn't be inserted into the database: ")
				break
			}
			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			if resp.StatusCode != http.StatusCreated {
				err := json.Unmarshal([]byte(body), &errorMessage)
				if err != nil {
					printErrorMessage(err, "Error parsing JSON: ")
					return
				}
				printMessage("Some error occurred: " + errorMessage.Message)
				break
			} else {
				printMessage("Availability added successfully!")

			}

		case "3":
			fmt.Println("Listing all availabilities for a specific teacher...")
			//retrieve data from cli for creating an availability
			name := getUserInput("Enter the teacher's name: ")
			surname := getUserInput("Enter the teacher's surname: ")
			teacher, err := getTeacherInfo(name, surname)
			if err != nil {
				printMessage("#### Impossible to retrieve the teacher's info ####")
				break
			}
			//api call
			baseUrl := fmt.Sprintf("http://localhost:8080/api/teacher/%d/availability", teacher.ID)
			resp, err := http.Get(baseUrl)
			if err != nil {
				break
			}
			//list out all the element of the rsponse body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			if string(body) == "[]" {
				printMessage("#### There are no availabilities for this teacher ####")
				break
			} else {
				var availabilities []Availability
				err = json.Unmarshal(body, &availabilities)
				if err != nil {
					printErrorMessage(err, "Error: ")
					break
				}
				fmt.Println("Availabilities: ")
				for i := 0; i < len(availabilities); i++ {
					fmt.Println("ID: ", availabilities[i].ID)
					fmt.Println("Day: ", availabilities[i].Day.Format("Monday, 2 January 2006"))
					fmt.Println("Starting time: ", availabilities[i].StartingTime.Format("15:04"))
					fmt.Println("Ending time: ", availabilities[i].EndingTime.Format("15:04"))
					fmt.Println("Booked: ", availabilities[i].Booked)
					fmt.Println("----------------------------------------------------------------")
				}
			}

		case "4":
			fmt.Println("Listing all teachers...")
			//api call
			url := "http://localhost:8080/api/teachers"

			resp, err := http.Get(url)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			//list out all the teachers
			if string(body) == "[]" {
				printMessage("#### There are no availabilities for this teacher ####")
				break
			} else {
				var teachers []Teacher
				err = json.Unmarshal(body, &teachers)
				if err != nil {
					printErrorMessage(err, "Error: ")
					break
				}
				fmt.Println("Teachers: ")
				for i := 0; i < len(teachers); i++ {
					fmt.Println("Teacher ID: ", teachers[i].ID)
					fmt.Println("Name: ", teachers[i].Name)
					fmt.Println("Surname: ", teachers[i].Surname)
					fmt.Println("----------------------------------------------------------------")
				}
			}
		case "5":
			fmt.Println("Adding a student...")
			//retrieve data from cli for creating a student
			name := getUserInput("Enter the student's name: ")
			surname := getUserInput("Enter the student's surname: ")

			date := getUserInput("Enter the student Date of Birth: ")
			dayStr := date[0:2]
			monthStr := date[3:5]
			yearStr := date[6:10]

			//convert string components to integers
			day, err := strconv.Atoi(dayStr)
			if err != nil {
				printErrorMessage(err, "Error converting day: ")
				break
			}

			month, err := strconv.Atoi(monthStr)
			if err != nil {
				printErrorMessage(err, "Error converting month: ")
				break
			}

			year, err := strconv.Atoi(yearStr)
			if err != nil {
				printErrorMessage(err, "Error converting year: ")
				break
			}
			if !isValidDate(year, month, day) {
				printMessage("Invalid date")
				break
			}
			//create a time.Time value using the components
			parsedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

			username := getUserInput("Enter the student's username: ")
			password := getUserInput("Enter the student's password: ")

			//create the object that represent the student
			student := Student{Name: name, Surname: surname, DateOfBirth: parsedDate, Username: username, Password: password}

			//api call
			url := "http://localhost:8080/api/student/addstudent"
			payload, err := json.Marshal(student)
			if err != nil {
				printMessage(err.Error())
				break
			}
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}

			if resp.StatusCode != http.StatusCreated {
				err := json.Unmarshal([]byte(body), &errorMessage)
				if err != nil {
					printErrorMessage(err, "Error parsing JSON: ")
					return
				}
				printMessage("Some error occurred: " + errorMessage.Message)
				break
			} else {
				printMessage("Student added successfully!")
			}
		case "6":
			fmt.Println("Listing all students...")
			//api call
			url := "http://localhost:8080/api/student/allstudents"

			resp, err := http.Get(url)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			//list out all the students
			if string(body) == "[]" {
				printMessage("#### There are no availabilities for this teacher ####")
				break
			} else {
				var students []Student
				err = json.Unmarshal(body, &students)
				if err != nil {
					printErrorMessage(err, "Error: ")
					break
				}
				fmt.Println("Students: ")
				for i := 0; i < len(students); i++ {
					fmt.Println("UserName: ", students[i].Username)
					fmt.Println("Name: ", students[i].Name)
					fmt.Println("Surname: ", students[i].Surname)
					fmt.Println("Date of birth: ", students[i].DateOfBirth.Format("Monday, 2 January 2006"))
					fmt.Println("----------------------------------------------------------------")
				}
			}
		case "7":
			fmt.Println("Showing profile of a specific student...")
			//retrieve data from cli for creating an availability
			username := getUserInput("Enter the student's username: ")
			//find it the username is already in use
			student, err := getStudentInfo(username)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			//retrieve password from the cli
			password := getUserInput("Enter the student's password: ")

			//check if the passowrd is correct
			err = bcrypt.CompareHashAndPassword([]byte(student.Password), []byte(password))
			if err != nil {
				printMessage("#### Wrong password ####")
				break
			}
			printStudentProfile(student)

		case "8":
			fmt.Println("Adding a booking for a specific teacher by the student X...")
			//retrieve username from the cli
			username := getUserInput("Enter the student's username: ")
			//retrieve ID of the student
			student, err := getStudentInfo(username)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			//retrieve form cli name and surname of the teacher
			teacherName := getUserInput("Enter the teacher's name: ")
			teacherSurname := getUserInput("Enter the teacher's surname: ")
			//retrieve ID of the teacher
			teacher, err := getTeacherInfo(teacherName, teacherSurname)
			if err != nil {
				printMessage("#### Couldn't get teacher information ####")
				break
			}
			//api call
			baseUrl := fmt.Sprintf("http://localhost:8080/api/teacher/%d/availability", teacher.ID)
			resp, err := http.Get(baseUrl)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			//list out all the element of the rsponse body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			var availabilities []Availability
			err = json.Unmarshal(body, &availabilities)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			count := 0
			if len(availabilities) != 0 {
				//parse into array of availabilities
				fmt.Println("Availabilties of " + teacherName + " " + teacherSurname)
				for _, a := range availabilities {
					if a.Booked == false {
						fmt.Printf("%d. %02d/%02d/%4d %02d:%02d - %02d:%02d\n",
							a.ID,
							a.Day.Day(),
							a.Day.Month(),
							a.Day.Year(),
							a.StartingTime.Hour(),
							a.StartingTime.Minute(),
							a.EndingTime.Hour(),
							a.EndingTime.Minute())
					} else {
						count = count + 1
					}
				}
				if count != len(availabilities) {
					//retrieve the ID of the availability
					fmt.Print("Enter the ID of the availability you want to book: ")
					var id int
					fmt.Scanln(&id)
					//select subject from the cli
					fmt.Print("Enter the subject you want to book: ")
					var subject string
					fmt.Scanln(&subject)

					var newBooking LessonReservation

					newBooking.StudentUsername = student.Username
					newBooking.TeacherID = teacher.ID
					newBooking.AvailabilityID = id
					newBooking.Subject = subject

					//api call
					url := "http://localhost:8080/api/student/" + student.Username + "/bookings"
					payload, err := json.Marshal(newBooking)
					resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
					if err != nil {
						printErrorMessage(err, "Error: ")
						break
					}
					defer resp.Body.Close()
					body, err = io.ReadAll(resp.Body)
					if err != nil {
						printErrorMessage(err, "Error reading response body: ")
						break
					}
					if resp.StatusCode != 201 {
						err = json.Unmarshal(body, &errorMessage)
						if err != nil {
							fmt.Println(err)
							break
						}
						printMessage("Some error occurred: " + errorMessage.Message)
						break
					} else {
						printMessage("Lesson booked successfully")
					}
				} else {
					printMessage("All the availabilities are already booked")
					break
				}

			} else {
				printMessage("There are no availabilities for this teacher")
				break
			}

		case "9":
			fmt.Println("Listing all the booking made by a specific student...")
			var student Student
			//retrieve username from cli
			student.Username = getUserInput("Enter the student's username: ")
			//api call 
			baseUrl := "http://localhost:8080/api/student/" + student.Username + "/bookings"
			resp, err := http.Get(baseUrl)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			defer resp.Body.Close()

			//list out all the booking
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				printErrorMessage(err, "Error reading response body: ")
				break
			}
			if resp.StatusCode != 200 {
				printMessage("#### No student found as " + student.Username + " ####")
				break
			}
			var bookings []LessonBooked
			err = json.Unmarshal(body, &bookings)
			if err != nil {
				printErrorMessage(err, "Error: ")
				break
			}
			for i := 0; i < len(bookings); i++ {
				fmt.Printf("%d. %s %02d:%02d - %02d:%02d - %s\n",
					bookings[i].ID,
					bookings[i].Day,
					bookings[i].StartingTime.Hour(),
					bookings[i].StartingTime.Minute(),
					bookings[i].EndingTime.Hour(),
					bookings[i].EndingTime.Minute(),
					bookings[i].Subject)
			}

		case "0":
			printMessage("Exiting the program. Goodbye!")
			os.Exit(0)

		default:
			printMessage("Invalid option. Please try again.")
		}
	}
}

func printMenu(test bool) {
	fmt.Println("\nMenu Options:")
	fmt.Println("1. Add a teacher")
	fmt.Println("2. Add an availability for a specific teacher")
	fmt.Println("3. List all availabilities for a specific teacher")
	fmt.Println("4. List all teachers")
	if test {
		fmt.Println("5. Add a student")
		fmt.Println("6. List all students")
		fmt.Println("7. Get profile of a specific student")
		fmt.Println("8. Book an availability for a specific teacher")
		fmt.Println("9. List all bookings for a specific student")
	}
	fmt.Println("0. Exit")
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func getStudentInfo(username string) (Student, error) { //used to retrieve the student from the username
	var student Student
	//api call 
	baseUrl := "http://localhost:8080/api/student/" + username + "/profile"
	resp, err := http.Get(baseUrl)
	if err != nil {
		printMessage("Error:" + err.Error())
		return Student{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printMessage("Error reading response body:" + err.Error())
		return Student{}, err
	}
	if resp.StatusCode != 200 {
		fmt.Println("#### No student found as " + username + " ####")
		return Student{}, errors.New("No student found")
	}
	err = json.Unmarshal(body, &student)
	if err != nil {
		printMessage(err.Error())
		return Student{}, err
	}
	return student, nil
}

func getTeacherInfo(teacherName, teacherSurname string) (Teacher, error) { //used to retrieve teacher info from name and surname
	var teacher Teacher
	//api call 
	baseUrl := "http://localhost:8080/api/teachers/" + teacherName + "/" + teacherSurname + "/"
	resp, err := http.Get(baseUrl)
	if err != nil {
		printMessage("Error:" + err.Error())
		return Teacher{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printMessage("Error reading response body:" + err.Error())
		return Teacher{}, err
	}
	if resp.StatusCode != 200 {
		printMessage("No teacher found as " + teacher.Name + " " + teacher.Surname)
		return Teacher{}, err
	}
	err = json.Unmarshal(body, &teacher)
	if err != nil {
		printMessage(err.Error())
		return Teacher{}, err
	}
	return teacher, nil
}

func printMessage(message string) {
	messageLength := len(message)
	topBottom := strings.Repeat("═", messageLength+2)
	sides := "║"

	// Print the top of the box
	fmt.Println("╔" + topBottom + "╗")

	// Print the message with padding
	fmt.Printf("%s %s %s\n", sides, message, sides)

	// Print the bottom of the box
	fmt.Println("╚" + topBottom + "╝")
}

func isValidDate(year, month, day int) bool { //check if the date is in the correct form
	// Construct a date string in the format "YYYY-MM-DD"
	dateString := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	// Parse the date string
	_, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		// Error parsing the date, it's not a valid date
		return false
	}

	// Check if the parsed date components match the provided year, month, and day
	return true
}

func printStudentProfile(student Student) {
	// Format date of birth
	dateOfBirth := student.DateOfBirth.Format("Monday, 2 January 2006")

	// Determine the maximum length among labels and values
	maxLength := len("Date of Birth:") + len(dateOfBirth)
	labels := []string{"Name:", "Surname:", "Date of Birth:", "Username:"}
	for _, label := range labels {
		if len(label) > maxLength {
			maxLength = len(label)
		}
	}
	maxLength = maxLength + 3
	// Create a formatted box with dynamic width
	box := fmt.Sprintf(`
╔%s╗
║ Student Profile %s║
╟%s╢
║ Name:          %s%s║
║ Surname:       %s%s║
║ Date of Birth: %s%s║
║ Username:      %s%s║
╚%s╝
`,
		strings.Repeat("═", maxLength),
		strings.Repeat(" ", maxLength-17),
		strings.Repeat("─", maxLength),
		student.Name, strings.Repeat(" ", maxLength-len("Name:        ")-len(student.Name)-3),
		student.Surname, strings.Repeat(" ", maxLength-len("Surname:     ")-len(student.Surname)-3),
		dateOfBirth, strings.Repeat(" ", maxLength-len("Date of Birth:")-len(dateOfBirth)-2),
		student.Username, strings.Repeat(" ", maxLength-len("Username:    ")-len(student.Username)-3),
		strings.Repeat("═", maxLength),
	)

	// Print the formatted box
	fmt.Println(box)
}

func printErrorMessage(err error, message ...string) {
	var errorMessage string

	if len(message) > 0 {
		errorMessage = fmt.Sprintf("%s %s", message[0], err.Error())
	} else {
		errorMessage = err.Error()
	}

	fmt.Printf("##### %s #####\n", errorMessage)
}
