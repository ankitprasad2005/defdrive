package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	step       int
	url        string
	token      string
	message    string
	input      string
	cursor     int
	showConfig bool
	menuIndex  int               // For arrow-key navigable menu
	signupData map[string]string // Store signup data
	loginData  map[string]string // Store login data
	fieldIndex int               // Track current field being edited
}

func (m Model) Init() tea.Cmd {
	// Check existing configuration
	url, token, username := loadConfig()
	if url != "" && token != "" {
		m.url = url
		m.token = token
		m.step = 2
		m.message = fmt.Sprintf("Existing configuration found:\nURL: %s\nUsername: %s\n\nDo you want to proceed with this? (yes/no):", url, username)
	} else if url != "" {
		m.url = url
		m.step = 0
		m.message = "Enter the URL where DefDrive is hosted:"
	} else {
		m.step = 0
		m.message = "Enter the URL where DefDrive is hosted:"
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.step {
			case 0: // Verify URL
				m.url = m.input
				saveConfig(m.url, "", "")
				m.step++
				m.message = "URL saved! Proceed to login or signup."
			case 1: // Handle menu selection
				if m.menuIndex == 0 { // Signup
					m.step = 3
					m.signupData = map[string]string{"name": "", "email": "", "username": "", "password": "", "confirm_password": ""}
					m.fieldIndex = 0
					m.message = "Enter your name:"
					m.input = ""
				} else if m.menuIndex == 1 { // Login
					m.step = 4
					m.loginData = map[string]string{"username": "", "password": ""}
					m.fieldIndex = 0
					m.message = "Enter your username:"
					m.input = ""
				}
			case 2: // Confirm existing config
				if strings.ToLower(m.input) == "yes" {
					m.step++
					m.message = "Proceeding with existing configuration..."
				} else if strings.ToLower(m.input) == "no" {
					m.url, m.token = "", ""
					m.step = 0
					m.message = "Resetting configuration. Please enter the URL."
				}
			case 3: // Signup input
				fields := []string{"name", "email", "username", "password", "confirm_password"}
				m.signupData[fields[m.fieldIndex]] = m.input
				m.input = ""
				m.fieldIndex++
				if m.fieldIndex < len(fields) {
					m.message = fmt.Sprintf("Enter your %s:", fields[m.fieldIndex])
				} else {
					m.step = 5
					m.message = fmt.Sprintf("Confirm your details:\nName: %s\nEmail: %s\nUsername: %s\nPassword: %s\nConfirm Password: %s\n\nType 'confirm' to proceed or 'edit' to modify.",
						m.signupData["name"], m.signupData["email"], m.signupData["username"], m.signupData["password"], m.signupData["confirm_password"])
				}
			case 4: // Login input
				fields := []string{"username", "password"}
				m.loginData[fields[m.fieldIndex]] = m.input
				m.input = ""
				m.fieldIndex++
				if m.fieldIndex < len(fields) {
					m.message = fmt.Sprintf("Enter your %s:", fields[m.fieldIndex])
				} else {
					m.step = 6
					m.message = fmt.Sprintf("Confirm your details:\nUsername: %s\nPassword: %s\n\nType 'confirm' to proceed or 'edit' to modify.",
						m.loginData["username"], m.loginData["password"])
				}
			case 5: // Confirm signup
				if strings.ToLower(m.input) == "confirm" {
					if m.signupData["password"] != m.signupData["confirm_password"] {
						m.step = 3
						m.fieldIndex = 3
						m.message = "Passwords do not match. Re-enter your password:"
					} else {
						token, err := signup(m.url, m.signupData["name"], m.signupData["email"], m.signupData["username"], m.signupData["password"])
						if err != nil {
							m.message = fmt.Sprintf("Signup failed: %v", err)
						} else {
							m.token = token
							saveConfig(m.url, m.token, m.signupData["username"])
							m.step++
							m.message = "Signup and login successful!"
						}
					}
				} else if strings.ToLower(m.input) == "edit" {
					m.step = 7
					m.fieldIndex = 0
					m.message = fmt.Sprintf("Modify your %s:", getSignupFields()[m.fieldIndex])
					m.input = m.signupData[getSignupFields()[m.fieldIndex]]
				}
			case 6: // Confirm login
				if strings.ToLower(m.input) == "confirm" {
					token, err := login(m.url, m.loginData["username"], m.loginData["password"])
					if err != nil {
						m.message = fmt.Sprintf("Login failed: %v", err)
					} else {
						m.token = token
						saveConfig(m.url, m.token, m.loginData["username"])
						m.step++
						m.message = "Login successful!"
					}
				} else if strings.ToLower(m.input) == "edit" {
					m.step = 8
					m.fieldIndex = 0
					m.message = fmt.Sprintf("Modify your %s:", getLoginFields()[m.fieldIndex])
					m.input = m.loginData[getLoginFields()[m.fieldIndex]]
				}
			case 7: // Modify signup fields
				fields := getSignupFields()
				m.signupData[fields[m.fieldIndex]] = m.input
				m.input = ""
				m.fieldIndex++
				if m.fieldIndex < len(fields) {
					m.message = fmt.Sprintf("Modify your %s:", fields[m.fieldIndex])
					m.input = m.signupData[fields[m.fieldIndex]]
				} else {
					m.step = 5
					m.message = fmt.Sprintf("Confirm your details:\nName: %s\nEmail: %s\nUsername: %s\nPassword: %s\nConfirm Password: %s\n\nType 'confirm' to proceed or 'edit' to modify.",
						m.signupData["name"], m.signupData["email"], m.signupData["username"], m.signupData["password"], m.signupData["confirm_password"])
				}
			case 8: // Modify login fields
				fields := getLoginFields()
				m.loginData[fields[m.fieldIndex]] = m.input
				m.input = ""
				m.fieldIndex++
				if m.fieldIndex < len(fields) {
					m.message = fmt.Sprintf("Modify your %s:", fields[m.fieldIndex])
					m.input = m.loginData[fields[m.fieldIndex]]
				} else {
					m.step = 6
					m.message = fmt.Sprintf("Confirm your details:\nUsername: %s\nPassword: %s\n\nType 'confirm' to proceed or 'edit' to modify.",
						m.loginData["username"], m.loginData["password"])
				}
			}
		case "up":
			if m.step == 1 && m.menuIndex > 0 {
				m.menuIndex--
			}
		case "down":
			if m.step == 1 && m.menuIndex < 1 {
				m.menuIndex++
			}
		case "backspace":
			if m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
		case "ctrl+shift+v":
			clipboardContent := "pasted_text" // Replace with actual clipboard handling if needed
			m.input = m.input[:m.cursor] + clipboardContent + m.input[m.cursor:]
			m.cursor += len(clipboardContent)
		default:
			if m.step == 0 || m.step == 2 || m.step == 3 || m.step == 4 || m.step == 5 || m.step == 6 || m.step == 7 || m.step == 8 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
			}
		}

		// Ensure the cursor is within valid bounds after updates
		if m.cursor < 0 {
			m.cursor = 0
		} else if m.cursor > len(m.input) {
			m.cursor = len(m.input)
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case 0:
		return fmt.Sprintf(
			"Enter the URL where DefDrive is hosted:\n%s\n\n%s\nPress Enter to continue or q to quit.",
			renderInput(m.input, m.cursor), m.message,
		)
	case 1:
		menu := []string{"Signup", "Login"}
		menuView := ""
		for i, item := range menu {
			cursor := " "
			if i == m.menuIndex {
				cursor = ">"
			}
			menuView += fmt.Sprintf("%s %s\n", cursor, item)
		}
		return fmt.Sprintf(
			"Select an option using the arrow keys and press Enter:\n\n%s\n\n%s",
			menuView, m.message,
		)
	case 2:
		return fmt.Sprintf(
			"Existing configuration found:\nURL: %s\nToken: %s\n\nDo you want to proceed with this? (yes/no):\n%s\n\n%s\nPress Enter to continue or q to quit.",
			m.url, m.token, renderInput(m.input, m.cursor), m.message,
		)
	case 3, 4, 5, 6, 7, 8:
		return fmt.Sprintf(
			"%s\n%s\nPress Enter to continue or q to quit.",
			m.message, renderInput(m.input, m.cursor),
		)
	default:
		return "Welcome to DefDrive CLI! Configuration complete. Press q to quit."
	}
}

func renderInput(input string, cursor int) string {
	// Ensure the cursor is within valid bounds
	if cursor < 0 {
		cursor = 0
	} else if cursor > len(input) {
		cursor = len(input)
	}
	return input[:cursor] + "|" + input[cursor:]
}

func main() {
	url, token, username := loadConfig() // Adjust to handle the third value
	model := Model{url: url, token: token, step: 2, showConfig: url != "" && token != ""}
	if !model.showConfig {
		model.step = 0
		model.message = "Enter the URL where DefDrive is hosted:"
	} else {
		model.message = fmt.Sprintf("Existing configuration found:\nURL: %s\nUsername: %s\n\nDo you want to proceed with this? (yes/no):", url, username)
	}
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting application: %v\n", err)
		os.Exit(1)
	}
}

func verifyURL(url string) error {
	resp, err := http.Get(url + "/api/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to connect to the server")
	}
	return nil
}

func signup(url, name, email, username, password string) (string, error) {
	requestBody := fmt.Sprintf(`{
		"name": "%s",
		"email": "%s",
		"username": "%s",
		"password": "%s"
	}`, name, email, username, password)

	resp, err := http.Post(url+"/api/signup", "application/json", strings.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("signup failed: %s", string(body))
	}

	// Automatically login after signup
	return login(url, username, password)
}

func login(url, username, password string) (string, error) {
	requestBody := fmt.Sprintf(`{
		"username": "%s",
		"password": "%s"
	}`, username, password)

	resp, err := http.Post(url+"/api/login", "application/json", strings.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed: %s", string(body))
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Token, nil
}

func saveConfig(url, token, username string) {
	config := map[string]string{"url": url, "token": token, "username": username}
	data, _ := json.Marshal(config)
	_ = ioutil.WriteFile(".defdrive_credentials", data, 0644)
}

func loadConfig() (string, string, string) {
	data, err := ioutil.ReadFile(".defdrive_credentials")
	if err != nil {
		return "", "", ""
	}
	var config map[string]string
	_ = json.Unmarshal(data, &config)
	return config["url"], config["token"], config["username"]
}

func getSignupFields() []string {
	return []string{"name", "email", "username", "password", "confirm_password"}
}

func getLoginFields() []string {
	return []string{"username", "password"}
}
