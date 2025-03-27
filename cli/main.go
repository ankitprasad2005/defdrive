package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state       string
	username    string
	password    string
	files       []File
	selectedIdx int
	accessLinks []Access
	inputBuffer string
	hostURL     string
	jwtToken    string
	cursorPos   int
	loginStep   string // Tracks whether the user is entering the username or password
	authStep    string // Tracks whether the user is choosing to sign up or log in
}

type File struct {
	ID       int    `json:"ID"`
	Name     string `json:"Name"`
	Location string `json:"Location"`
	Size     int64  `json:"Size"`
	Public   bool   `json:"Public"`
}

type Access struct {
	ID   int    `json:"ID"`
	Name string `json:"Name"`
	Link string `json:"Link"`
}

func initialModel() model {
	// Check if host URL and JWT token exist
	hostURL, jwtToken := loadCredentials()
	if hostURL != "" && jwtToken != "" {
		return model{
			state:    "connecting",
			hostURL:  hostURL,
			jwtToken: jwtToken,
		}
	}
	return model{
		state: "host",
	}
}

func loadCredentials() (string, string) {
	data, err := ioutil.ReadFile(".defdrive_credentials")
	if err != nil {
		return "", ""
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return "", ""
	}
	return lines[0], lines[1]
}

func saveCredentials(hostURL, jwtToken string) {
	data := fmt.Sprintf("%s\n%s", hostURL, jwtToken)
	_ = ioutil.WriteFile(".defdrive_credentials", []byte(data), 0600)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) login() error {
	url := fmt.Sprintf("%s/api/login", m.hostURL)
	payload := map[string]string{
		"username": m.username,
		"password": m.password,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to login: %s", resp.Status)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	m.jwtToken = result.Token
	saveCredentials(m.hostURL, m.jwtToken)
	return nil
}

func (m *model) signup() error {
	url := fmt.Sprintf("%s/api/signup", m.hostURL)
	payload := map[string]string{
		"name":     m.username,                                // Using username as name for simplicity
		"email":    fmt.Sprintf("%s@example.com", m.username), // Dummy email
		"username": m.username,
		"password": m.password,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sign up: %s", resp.Status)
	}

	return nil
}

func (m *model) fetchFiles() error {
	url := fmt.Sprintf("%s/api/files", m.hostURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.jwtToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch files: %s", resp.Status)
	}

	var result struct {
		Files []File `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	m.files = result.Files
	return nil
}

func (m *model) fetchAccessLinks(fileID int) error {
	url := fmt.Sprintf("%s/api/files/%d/accesses", m.hostURL, fileID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.jwtToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch access links: %s", resp.Status)
	}

	var result struct {
		Accesses []Access `json:"accesses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	m.accessLinks = result.Accesses
	return nil
}

func (m *model) uploadFile(filePath string) error {
	url := fmt.Sprintf("%s/api/upload", m.hostURL)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy the file content to the form
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}
	writer.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.jwtToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return nil
}

func (m *model) addAccessLink(fileID int) error {
	url := fmt.Sprintf("%s/api/files/%d/accesses", m.hostURL, fileID)

	// Prompt user for access link parameters
	fmt.Println("Enter the following details for the new access link:")
	fmt.Print("Name: ")
	var name string
	fmt.Scanln(&name)

	fmt.Print("Subnets (comma-separated, leave empty for none): ")
	var subnetsInput string
	fmt.Scanln(&subnetsInput)
	subnets := strings.Split(subnetsInput, ",")

	fmt.Print("IPs (comma-separated, leave empty for none): ")
	var ipsInput string
	fmt.Scanln(&ipsInput)
	ips := strings.Split(ipsInput, ",")

	fmt.Print("Expires (ISO 8601 format, leave empty for no expiration): ")
	var expires string
	fmt.Scanln(&expires)

	fmt.Print("Public (true/false): ")
	var public bool
	fmt.Scanln(&public)

	fmt.Print("One-Time Use (true/false): ")
	var oneTimeUse bool
	fmt.Scanln(&oneTimeUse)

	fmt.Print("TTL (integer, leave empty for none): ")
	var ttl int
	fmt.Scanln(&ttl)

	fmt.Print("Enable TTL (true/false): ")
	var enableTTL bool
	fmt.Scanln(&enableTTL)

	payload := map[string]interface{}{
		"name":       name,
		"subnets":    subnets,
		"ips":        ips,
		"expires":    expires,
		"public":     public,
		"oneTimeUse": oneTimeUse,
		"ttl":        ttl,
		"enableTTL":  enableTTL,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+m.jwtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create access link: %s", resp.Status)
	}

	return nil
}

func validateHostURL(input string) (string, error) {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", fmt.Errorf("invalid URL: must include http:// or https://")
	}
	return parsedURL.String(), nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "connecting":
			if err := m.fetchFiles(); err != nil {
				log.Printf("Failed to connect: %v", err)
				m.state = "host"
				m.inputBuffer = ""
			} else {
				m.state = "main"
			}
		case "host", "signup", "login", "upload":
			switch msg.String() {
			case "enter":
				// Handle enter key logic for each state
				switch m.state {
				case "host":
					trimmedInput := strings.TrimSpace(m.inputBuffer)
					validatedURL, err := validateHostURL(trimmedInput)
					if err != nil {
						log.Printf("Invalid host URL: %v", err)
						m.inputBuffer = "" // Clear the input buffer for retry
						return m, nil
					}
					m.hostURL = validatedURL
					m.state = "auth"
					m.inputBuffer = ""
				case "signup":
					parts := strings.SplitN(m.inputBuffer, " ", 2)
					if len(parts) != 2 {
						log.Println("Invalid input. Enter username and password separated by a space.")
						return m, nil
					}
					m.username = parts[0]
					m.password = parts[1]

					if err := m.signup(); err != nil {
						log.Printf("Signup failed: %v", err)
						return m, nil
					}

					log.Println("Signup successful. Proceeding to login.")
					m.state = "login"
					m.inputBuffer = ""
					m.loginStep = "username"
				case "login":
					if m.loginStep == "username" {
						m.username = strings.TrimSpace(m.inputBuffer)
						m.inputBuffer = ""
						m.loginStep = "password"
					} else if m.loginStep == "password" {
						m.password = strings.TrimSpace(m.inputBuffer)
						m.inputBuffer = ""

						if err := m.login(); err != nil {
							log.Printf("Login failed: %v", err)
							m.loginStep = "username"
							return m, nil
						}

						if err := m.fetchFiles(); err != nil {
							log.Printf("Failed to fetch files: %v", err)
							m.loginStep = "username"
							return m, nil
						}

						m.state = "main"
					}
				case "upload":
					filePath := strings.TrimSpace(m.inputBuffer)
					if filePath == "" {
						log.Println("Invalid file path.")
						return m, nil
					}

					if err := m.uploadFile(filePath); err != nil {
						log.Printf("Failed to upload file: %v", err)
						return m, nil
					}

					log.Println("File uploaded successfully.")
					if err := m.fetchFiles(); err != nil {
						log.Printf("Failed to refresh files: %v", err)
						return m, nil
					}

					m.state = "main"
					m.inputBuffer = ""
				}
			case "backspace":
				if m.cursorPos > 0 && len(m.inputBuffer) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.inputBuffer[:m.cursorPos])
					m.inputBuffer = m.inputBuffer[:m.cursorPos-size] + m.inputBuffer[m.cursorPos:]
					m.cursorPos -= size
				}
			case "left":
				if m.cursorPos > 0 {
					_, size := utf8.DecodeLastRuneInString(m.inputBuffer[:m.cursorPos])
					m.cursorPos -= size
				}
			case "right":
				if m.cursorPos < len(m.inputBuffer) {
					_, size := utf8.DecodeRuneInString(m.inputBuffer[m.cursorPos:])
					m.cursorPos += size
				}
			case "paste":
				// Handle paste input
				m.inputBuffer = m.inputBuffer[:m.cursorPos] + msg.String() + m.inputBuffer[m.cursorPos:]
				m.cursorPos += len(msg.String())
			default:
				if len(msg.String()) == 1 {
					m.inputBuffer = m.inputBuffer[:m.cursorPos] + msg.String() + m.inputBuffer[m.cursorPos:]
					m.cursorPos += len(msg.String())
				}
			}
		case "auth":
			if msg.String() == "1" {
				m.state = "signup"
				m.inputBuffer = ""
			} else if msg.String() == "2" {
				m.state = "login"
				m.inputBuffer = ""
				m.loginStep = "username"
			} else if msg.String() == "q" {
				return m, tea.Quit
			}
		case "main":
			if msg.String() == "q" {
				return m, tea.Quit
			} else if msg.String() == "enter" {
				if len(m.files) > 0 {
					selectedFile := m.files[m.selectedIdx]
					if err := m.fetchAccessLinks(selectedFile.ID); err != nil {
						log.Printf("Failed to fetch access links: %v", err)
						return m, nil
					}
					m.state = "file"
				}
			} else if msg.String() == "u" {
				m.state = "upload"
				m.inputBuffer = ""
			} else if msg.String() == "up" {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			} else if msg.String() == "down" {
				if m.selectedIdx < len(m.files)-1 {
					m.selectedIdx++
				}
			}
		case "file":
			if msg.String() == "q" {
				m.state = "main"
			} else if msg.String() == "a" {
				if len(m.files) > 0 {
					selectedFile := m.files[m.selectedIdx]
					if err := m.addAccessLink(selectedFile.ID); err != nil {
						log.Printf("Failed to add access link: %v", err)
						return m, nil
					}
					log.Println("Access link added successfully.")
					if err := m.fetchAccessLinks(selectedFile.ID); err != nil {
						log.Printf("Failed to refresh access links: %v", err)
					}
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case "connecting":
		return fmt.Sprintf("Connecting to %s...\n", m.hostURL)
	case "host":
		return fmt.Sprintf(
			"Enter the URL where Defdrive is hosted (e.g., http://localhost:5050): %s\nPress Enter to continue or q to exit",
			renderInputWithCursor(m.inputBuffer, m.cursorPos),
		)
	case "auth":
		return "Choose an option:\n1. Sign Up\n2. Log In\nq. Exit\nEnter your choice: "
	case "signup":
		return fmt.Sprintf(
			"Enter your username and password (space-separated): %s\nPress Enter to continue or q to go back",
			renderInputWithCursor(m.inputBuffer, m.cursorPos),
		)
	case "login":
		if m.loginStep == "username" {
			return fmt.Sprintf(
				"Enter your username: %s\nPress Enter to continue or q to go back",
				renderInputWithCursor(m.inputBuffer, m.cursorPos),
			)
		} else if m.loginStep == "password" {
			return fmt.Sprintf(
				"Enter your password: %s\nPress Enter to continue or q to go back",
				renderInputWithCursor(strings.Repeat("*", utf8.RuneCountInString(m.inputBuffer)), m.cursorPos),
			)
		}
	case "main":
		return fmt.Sprintf("Files:\n%s\nUse Up/Down to navigate, Enter to select, u to upload a file, q to quit", m.renderFiles())
	case "upload":
		return fmt.Sprintf(
			"Enter the file path to upload: %s\nPress Enter to upload or q to go back",
			renderInputWithCursor(m.inputBuffer, m.cursorPos),
		)
	case "file":
		return fmt.Sprintf("Access Links:\n%s\nPress a to add a new access link, q to go back", m.renderAccessLinks())
	}
	return ""
}

func renderInputWithCursor(input string, cursorPos int) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(
		input[:cursorPos] + "|" + input[cursorPos:],
	)
}

func (m model) renderFiles() string {
	var result string
	for i, file := range m.files {
		cursor := " "
		if i == m.selectedIdx {
			cursor = ">"
		}
		result += fmt.Sprintf("%s %d. %s (Size: %d, Public: %t)\n", cursor, i+1, file.Name, file.Size, file.Public)
	}
	return result
}

func (m model) renderAccessLinks() string {
	var result string
	for i, access := range m.accessLinks {
		result += fmt.Sprintf("%d. %s (Link: %s)\n", i+1, access.Name, access.Link)
	}
	return result
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatalf("Error starting program: %v", err)
		os.Exit(1)
	}
}
