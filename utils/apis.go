package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// --- Piston (Code Runner) ---
type PistonRequest struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Files    []struct {
		Content string `json:"content"`
	} `json:"files"`
}

type PistonResponse struct {
	Run struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Output string `json:"output"`
	} `json:"run"`
}

func RunCode(lang, code string) string {
	versions := map[string]string{
		"python": "3.10.0", "js": "18.15.0", "javascript": "18.15.0",
		"go": "1.18.1", "cpp": "10.2.0", "java": "15.0.2",
		"rust": "1.68.2", "php": "8.2.3", "ruby": "3.0.1",
	}

	ver, ok := versions[strings.ToLower(lang)]
	if !ok { return "Unsupported language." }

	reqBody := PistonRequest{
		Language: lang, Version: ver,
		Files: []struct{ Content string `json:"content"` }{{Content: code}},
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("https://emkc.org/api/v2/piston/execute", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil { return "Error contacting code execution engine." }
	defer resp.Body.Close()

	var pistonResp PistonResponse
	json.NewDecoder(resp.Body).Decode(&pistonResp)

	output := pistonResp.Run.Output
	if output == "" { output = "No output." }
	if len(output) > 1000 { output = output[:1000] + "\n... (truncated)" }

	return fmt.Sprintf("💻 *Result (%s):*\n```\n%s\n```", lang, output)
}

// --- Jikan (Anime) ---
func GetAnimeInfo(query string) string {
	url := fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s&limit=1", query)
	resp, err := http.Get(url)
	if err != nil { return "Error fetching anime info." }
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			Title    string `json:"title"`
			Score    float64 `json:"score"`
			Synopsis string `json:"synopsis"`
			Status   string `json:"status"`
			Episodes int    `json:"episodes"`
			URL      string `json:"url"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if len(data.Data) == 0 { return "Anime not found." }
	a := data.Data[0]
	synopsis := a.Synopsis
	if len(synopsis) > 300 { synopsis = synopsis[:300] + "..." }

	return fmt.Sprintf("🎥 *%s*\n⭐ Score: %.2f\n📺 Status: %s (%d eps)\n📝 Synopsis: %s\n🔗 Link: %s", a.Title, a.Score, a.Status, a.Episodes, synopsis, a.URL)
}

// --- GitHub ---
func GetGithubProfile(username string) string {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	resp, err := http.Get(url)
	if err != nil { return "Error fetching GitHub profile." }
	defer resp.Body.Close()

	var data struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		PublicRepos int  `json:"public_repos"`
		Followers   int  `json:"followers"`
		Following   int  `json:"following"`
		HTMLURL     string `json:"html_url"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if data.Login == "" { return "GitHub user not found." }
	return fmt.Sprintf("🐙 *GitHub Profile: %s*\n👤 Name: %s\n📝 Bio: %s\n📂 Repos: %d | 👥 Followers: %d | 👤 Following: %d\n🔗 URL: %s", data.Login, data.Name, data.Bio, data.PublicRepos, data.Followers, data.Following, data.HTMLURL)
}

func GetGithubRepo(repoPath string) string {
	url := fmt.Sprintf("https://api.github.com/repos/%s", repoPath)
	resp, err := http.Get(url)
	if err != nil { return "Error fetching GitHub repo." }
	defer resp.Body.Close()

	var data struct {
		FullName        string `json:"full_name"`
		Description     string `json:"description"`
		StargazersCount int    `json:"stargazers_count"`
		ForksCount      int    `json:"forks_count"`
		Language        string `json:"language"`
		HTMLURL         string `json:"html_url"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if data.FullName == "" { return "GitHub repo not found." }
	return fmt.Sprintf("📁 *GitHub Repo: %s*\n⭐ Stars: %d | 🍴 Forks: %d\n💻 Language: %s\n📝 Desc: %s\n🔗 URL: %s", data.FullName, data.StargazersCount, data.ForksCount, data.Language, data.Description, data.HTMLURL)
}
