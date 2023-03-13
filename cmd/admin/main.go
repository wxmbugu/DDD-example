package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/patienttracker/internal/services"
	"golang.org/x/term"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var cliname string = "admin"

func printPrompt() {
	fmt.Print(cliname, "> ")
}

func printUnknown(text string) {
	fmt.Println(text, ": command not found!!")
}

func printError(err string) {
	fmt.Println("An error occurred: ", err)
}

func displayHelp() {
	fmt.Printf(
		"Welcome to %v! These are the available commands: \n",
		cliname,
	)
	fmt.Println("help    - Show available commands")
	fmt.Println("createadmin  - create admin")
	fmt.Println("createroles  - create roles")
	fmt.Println("clear   - Clear the terminal screen")
	fmt.Println("exit    - exit out of the prompt ")

}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func handleInvalidCmd(text string) {
	defer printUnknown(text)
}

func handleError(error string) {
	defer printError(error)
}
func handleCmd(text string) {
	handleInvalidCmd(text)
}

func cleanInput(text string) string {
	output := strings.TrimSpace(text)
	output = strings.ToLower(output)
	return output
}

func createAdmin() {
	fmt.Println("Enter Your Email: ")
	var email string
	var err error
	// Taking input from user
	fmt.Scanln(&email)
	if _, err = mail.ParseAddress(email); err != nil {
		handleError(err.Error())
		createAdmin()
	}
	fmt.Println("Enter Your Password: ")
	var password string
	// fmt.Scanln(&password)
	pass, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		handleError(err.Error())
		createAdmin()
	}
	password = string(pass)
	fmt.Println("Confirm Password: ")
	var confirmpassword string
	// fmt.Scanln(&password)
	cpass, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		handleError(err.Error())
		createAdmin()
	}
	confirmpassword = string(cpass)
	if password != confirmpassword {
		handleError("passwords don't match")
		createAdmin()
	} else {
		conn := SetupDb("postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable")
		service := services.NewService(conn)
		if _, err := service.CreateAdmin(email, password); err != nil {
			handleError(err.Error())
			createAdmin()
		}
		fmt.Println("admin created")
	}
}

func SetupDb(conn string) *sql.DB {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	db.SetConnMaxLifetime(time.Hour)
	return db
}

// TODO: Create ROLES
func main() {
	// Hardcoded repl commands
	commands := map[string]interface{}{
		"help":        displayHelp,
		"clear":       clearScreen,
		"createadmin": createAdmin,
	}
	// Begin the repl loop
	reader := bufio.NewScanner(os.Stdin)
	printPrompt()
	for reader.Scan() {
		text := cleanInput(reader.Text())
		if command, exists := commands[text]; exists {
			command.(func())()
		} else if strings.EqualFold("exit", text) {
			return
		} else {
			handleCmd(text)
		}
		printPrompt()
	}
	// Print an additional line if we encountered an EOF character
	fmt.Println()
}
