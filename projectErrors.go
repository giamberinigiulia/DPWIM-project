package main

import "fmt"

type ErrTeacherNotFound struct {
	TeacherID int
}
type ErrStudentNotFound struct {
	StudentID string
}

func (e *ErrTeacherNotFound) Error() string {
	return fmt.Sprintf("No Teacher with id: %d", e.TeacherID)
}

func (e *ErrStudentNotFound) Error() string {
	return fmt.Sprintf("No Student with username: %s", e.StudentID)
}
