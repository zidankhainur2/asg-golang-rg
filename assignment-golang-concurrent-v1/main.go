package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"a21hc3NpZ25tZW50/helper"
	"a21hc3NpZ25tZW50/model"
)

type StudentManager interface {
	Login(id string, name string) error
	Register(id string, name string, studyProgram string) error
	GetStudyProgram(code string) (string, error)
	ModifyStudent(name string, fn model.StudentModifier) error
}

type InMemoryStudentManager struct {
	sync.Mutex
	students             []model.Student
	studentStudyPrograms map[string]string
	failedLoginAttempts  map[string]int
}

func NewInMemoryStudentManager() *InMemoryStudentManager {
	return &InMemoryStudentManager{
		students: []model.Student{
			{
				ID:           "A12345",
				Name:         "Aditira",
				StudyProgram: "TI",
			},
			{
				ID:           "B21313",
				Name:         "Dito",
				StudyProgram: "TK",
			},
			{
				ID:           "A34555",
				Name:         "Afis",
				StudyProgram: "MI",
			},
		},
		studentStudyPrograms: map[string]string{
			"TI": "Teknik Informatika",
			"TK": "Teknik Komputer",
			"SI": "Sistem Informasi",
			"MI": "Manajemen Informasi",
		},
		failedLoginAttempts: make(map[string]int),
	}
}

func ReadStudentsFromCSV(filename string) ([]model.Student, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3 // ID, Name and StudyProgram

	var students []model.Student
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		student := model.Student{
			ID:           record[0],
			Name:         record[1],
			StudyProgram: record[2],
		}
		students = append(students, student)
	}
	return students, nil
}

func (sm *InMemoryStudentManager) GetStudents() []model.Student {
	return sm.students
}

func (sm *InMemoryStudentManager) Login(id string, name string) (string, error) {
	if id == "" || name == "" {
		return "", fmt.Errorf("ID or Name is undefined")
	}

	if sm.failedLoginAttempts[id] >= 3 {
		return "", fmt.Errorf("Login gagal: Batas maksimum login terlampaui")
	}

	for _, student := range sm.students {
		if student.ID == id && student.Name == name {
			sm.failedLoginAttempts[id] = 0

			studyProgram, _ := sm.GetStudyProgram(student.StudyProgram)
			return fmt.Sprintf("Login berhasil: Selamat datang %s! Kamu terdaftar di program studi: %s", name, studyProgram), nil
		}
	}
	sm.failedLoginAttempts[id]++
	return "", fmt.Errorf("Login gagal: data mahasiswa tidak ditemukan")
}

func (sm *InMemoryStudentManager) RegisterLongProcess() {
	// 30ms delay to simulate slow processing
	time.Sleep(30 * time.Millisecond)
}

func (sm *InMemoryStudentManager) Register(id string, name string, studyProgram string) (string, error) {
	// 30ms delay to simulate slow processing. DO NOT REMOVE THIS LINE
	sm.RegisterLongProcess()

	// Below lock is needed to prevent data race error. DO NOT REMOVE BELOW 2 LINES
	sm.Lock()
	defer sm.Unlock()

	if id == "" || name == "" || studyProgram == "" {
		return "", fmt.Errorf("ID, Name or StudyProgram is undefined!")
	}

	if _, exist := sm.studentStudyPrograms[studyProgram]; !exist {
		return "", fmt.Errorf("Study program %s is not found", studyProgram)
	}

	for _, student := range sm.students {
		if student.ID == id {
			return "", fmt.Errorf("Registrasi gagal: id sudah digunakan")
		}
	}

	sm.students = append(sm.students, model.Student{
		ID:           id,
		Name:         name,
		StudyProgram: studyProgram,
	})
	return fmt.Sprintf("Registrasi berhasil: %s (%s)", name, studyProgram), nil
}

func (sm *InMemoryStudentManager) GetStudyProgram(code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("Code is undefined!")
	}

	studyProgram, exist := sm.studentStudyPrograms[code]
	if !exist {
		return "", fmt.Errorf("Kode program studi tidak ditemukan")
	}

	return studyProgram, nil
}

func (sm *InMemoryStudentManager) ModifyStudent(name string, fn model.StudentModifier) (string, error) {
	for i, student := range sm.students {
		if student.Name == name {
			err := fn(&sm.students[i])
			if err != nil {
				return "", err
			}
			return "Program studi mahasiswa berhasil diubah.", nil
		}
	}
	return "", fmt.Errorf("Mahasiswa tidak ditemukan.")
}

func (sm *InMemoryStudentManager) ChangeStudyProgram(programStudi string) model.StudentModifier {
	if _, exists := sm.studentStudyPrograms[programStudi]; !exists {
		return func(s *model.Student) error {
			return fmt.Errorf("Kode program studi tidak ditemukan")
		}
	}

	return func(s *model.Student) error {
		s.StudyProgram = programStudi
		return nil
	}
}

func (sm *InMemoryStudentManager) ImportStudents(filenames []string) error {
	var wg sync.WaitGroup
	ch := make(chan []model.Student, len(filenames))
	errCh := make(chan error, len(filenames))

	maxConcurrent := 5
	sem := make(chan struct{}, maxConcurrent)

	for _, filename := range filenames {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			sem <- struct{}{}
			students, err := ReadStudentsFromCSV(file)
			if err != nil {
				errCh <- err
				<-sem
				return
			}
			ch <- students
			<-sem
		}(filename)
	}

	go func() {
		wg.Wait()
		close(ch)
		close(errCh)
	}()

	var registerWg sync.WaitGroup
	for students := range ch {
		for _, student := range students {
			registerWg.Add(1)
			go func(st model.Student) {
				defer registerWg.Done()
				if _, err := sm.Register(st.ID, st.Name, st.StudyProgram); err != nil {
					errCh <- err
				}
			}(student)
		}
	}
	registerWg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	return nil
}

func (sm *InMemoryStudentManager) SubmitAssignmentLongProcess() {
	// 3000ms delay to simulate slow processing
	time.Sleep(30 * time.Millisecond)
}

func (sm *InMemoryStudentManager) SubmitAssignments(numAssignments int) {

	start := time.Now()

	jobQueue := make(chan int, numAssignments)

	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(WorkerId int) {
			defer wg.Done()
			for assignment := range jobQueue {
				fmt.Printf("Worker %d submitted assignment %d\n", WorkerId, assignment)
				sm.SubmitAssignmentLongProcess()
				fmt.Printf("Worker %d submitted assignment %d\n", WorkerId, assignment)
			}
		}(i)
	}

	for i := 1; i <= numAssignments; i++ {
		jobQueue <- i
	}

	close(jobQueue)

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Submitting %d assignments took %s\n", numAssignments, elapsed)
}

func main() {
	manager := NewInMemoryStudentManager()

	for {
		helper.ClearScreen()
		students := manager.GetStudents()
		for _, student := range students {
			fmt.Printf("ID: %s\n", student.ID)
			fmt.Printf("Name: %s\n", student.Name)
			fmt.Printf("Study Program: %s\n", student.StudyProgram)
			fmt.Println()
		}

		fmt.Println("Selamat datang di Student Portal!")
		fmt.Println("1. Login")
		fmt.Println("2. Register")
		fmt.Println("3. Get Study Program")
		fmt.Println("4. Modify Student")
		fmt.Println("5. Bulk Import Student")
		fmt.Println("6. Submit assignment")
		fmt.Println("7. Exit")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Pilih menu: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			helper.ClearScreen()
			fmt.Println("=== Login ===")
			fmt.Print("ID: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)

			fmt.Print("Name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			msg, err := manager.Login(id, name)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			}
			fmt.Println(msg)
			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')
		case "2":
			helper.ClearScreen()
			fmt.Println("=== Register ===")
			fmt.Print("ID: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)

			fmt.Print("Name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Study Program Code (TI/TK/SI/MI): ")
			code, _ := reader.ReadString('\n')
			code = strings.TrimSpace(code)

			msg, err := manager.Register(id, name, code)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			}
			fmt.Println(msg)
			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')
		case "3":
			helper.ClearScreen()
			fmt.Println("=== Get Study Program ===")
			fmt.Print("Program Code (TI/TK/SI/MI): ")
			code, _ := reader.ReadString('\n')
			code = strings.TrimSpace(code)

			if studyProgram, err := manager.GetStudyProgram(code); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				fmt.Printf("Program Studi: %s\n", studyProgram)
			}
			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')
		case "4":
			helper.ClearScreen()
			fmt.Println("=== Modify Student ===")
			fmt.Print("Name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Program Studi Baru (TI/TK/SI/MI): ")
			code, _ := reader.ReadString('\n')
			code = strings.TrimSpace(code)

			msg, err := manager.ModifyStudent(name, manager.ChangeStudyProgram(code))
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			}
			fmt.Println(msg)

			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')
		case "5":
			helper.ClearScreen()
			fmt.Println("=== Bulk Import Student ===")

			// Define the list of CSV file names
			csvFiles := []string{"students1.csv", "students2.csv", "students3.csv"}

			err := manager.ImportStudents(csvFiles)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				fmt.Println("Import successful!")
			}

			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')

		case "6":
			helper.ClearScreen()
			fmt.Println("=== Submit Assignment ===")

			// Enter how many assignments you want to submit
			fmt.Print("Enter the number of assignments you want to submit: ")
			numAssignments, _ := reader.ReadString('\n')

			// Convert the input to an integer
			numAssignments = strings.TrimSpace(numAssignments)
			numAssignmentsInt, err := strconv.Atoi(numAssignments)

			if err != nil {
				fmt.Println("Error: Please enter a valid number")
			}

			manager.SubmitAssignments(numAssignmentsInt)

			// Wait until the user presses any key
			fmt.Println("Press any key to continue...")
			reader.ReadString('\n')
		case "7":
			helper.ClearScreen()
			fmt.Println("Goodbye!")
			return
		default:
			helper.ClearScreen()
			fmt.Println("Pilihan tidak valid!")
			helper.Delay(5)
		}

		fmt.Println()
	}
}
