# Tutoring Web App: GoTutor

## Overview

GoTutor is a tutoring web application designed to facilitate online tutoring sessions between students and tutors. It provides a user-friendly platform for students to prenote lessons. The application is built using GoLang for the backend server and SQLite for data storage.

## Features

- User-friendly interface for browsing available lessons and scheduling appointments.
- Secure authentication and authorization mechanisms for user accounts.
- CRUD operations for managing student and tutor profiles, lesson prenotations, and availability slots.

## Getting Started

To test the project, follow these steps:

1. **Download the project zip**: Download the project zip file from the GitHub repository.

2. **Install GoLang**: Ensure that GoLang is installed on your system. You can download and install it from the official GoLang website: [GoLang Website](https://golang.org/)

3. **Extract the project files**: Extract the contents of the downloaded zip file to a directory on your local machine.

4. **Navigate to the project directory**: Open a terminal or command prompt and navigate to the directory where you extracted the project files.

5. **Build the project**: Use the appropriate command to run the project:

   ```bash
   go build -o server.exe
   
6. **Run the project**:
      ```bash
   server.exe -m server //for launching the API server and the database

   server.exe -m web //for launching the Web Server

   server.exe -m cli //for launching the CLI interface

   server.exe -m cli -test //for testing all the teacher and student-related operations
