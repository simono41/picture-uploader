package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	lastUploadTime time.Time
	mu             sync.Mutex
	uploadInterval = 10 * time.Second
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/view/", viewHandler)

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Setzen der Content Security Policy
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")

	// Verwenden von html/template zur sicheren Ausgabe von HTML
	tmpl, err := template.ParseFiles("templates/homeTemplate.html")
	if err != nil {
		http.Error(w, "Fehler beim Laden des Templates", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Bildupload",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Fehler beim Rendern des Templates", http.StatusInternalServerError)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(lastUploadTime) < uploadInterval {
		http.Error(w, "Nur alle 10 Sekunden erlaubt", http.StatusTooManyRequests)
		log.Printf("Bildupload zu häufig. Nur alle 10 Sekunden erlaubt.")
		return
	}

	if r.Method == http.MethodPost {
		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Fehler beim Lesen der Datei", http.StatusInternalServerError)
			log.Printf("Fehler beim Lesen der Datei: %v", err)
			return
		}
		defer file.Close()

		// Überprüfen Sie den MIME-Typ der Datei
		buffer := make([]byte, 512) // Genug für die Erkennung des MIME-Typs
		_, err = file.Read(buffer)
		if err != nil {
			http.Error(w, "Fehler beim Lesen der Datei", http.StatusInternalServerError)
			log.Printf("Fehler beim Lesen der Datei für MIME-Typ-Erkennung: %v", err)
			return
		}

		mimeType := http.DetectContentType(buffer)
		if !strings.HasPrefix(mimeType, "image/") {
			http.Error(w, "Nur Bild-Uploads sind erlaubt", http.StatusBadRequest)
			log.Printf("Versuch, eine Nicht-Bild-Datei hochzuladen: %v", mimeType)
			return
		}

		// Zurücksetzen des Dateizeigers, um die Datei vollständig zu kopieren
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, "Fehler beim Zurücksetzen des Dateizeigers", http.StatusInternalServerError)
			log.Printf("Fehler beim Zurücksetzen des Dateizeigers: %v", err)
			return
		}

		// Hier können Sie den Dateinamen manipulieren oder einen anderen Speicherort wählen
		uploadPath := "./uploads/" + handler.Filename
		f, err := os.Create(uploadPath)
		if err != nil {
			http.Error(w, "Fehler beim Erstellen der Datei", http.StatusInternalServerError)
			log.Printf("Fehler beim Erstellen der Datei: %v", err)
			return
		}
		defer f.Close()

		_, copyErr := io.Copy(f, file)
		if copyErr != nil {
			http.Error(w, "Fehler beim Kopieren der Datei", http.StatusInternalServerError)
			log.Printf("Fehler beim Kopieren der Datei: %v", copyErr)
			return
		}

		lastUploadTime = time.Now() // Setzen Sie die Zeit des letzten Uploads

		// Nach erfolgreichem Upload:
		tmpl, err := template.ParseFiles("templates/uploadSuccess.html")
		if err != nil {
			http.Error(w, "Fehler beim Laden des Templates", http.StatusInternalServerError)
			log.Printf("Fehler beim Laden des Templates: %v", err)
			return
		}

		data := struct {
			Message  string
			Filename string
		}{
			Message:  "Bild erfolgreich hochgeladen.",
			Filename: handler.Filename,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Fehler beim Rendern des Templates", http.StatusInternalServerError)
			log.Printf("Fehler beim Rendern des Templates: %v", err)
			return
		}

	} else {

		tmpl, err := template.ParseFiles("templates/uploadForm.html")
		if err != nil {
			http.Error(w, "Fehler beim Laden des Templates", http.StatusInternalServerError)
			log.Printf("Fehler beim Laden des Templates: %v", err)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Fehler beim Rendern des Templates", http.StatusInternalServerError)
			log.Printf("Fehler beim Rendern des Templates: %v", err)
			return
		}
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len("/view/"):]
	imagePath := "./uploads/" + filePath

	// Öffnen und Lesen der Bilddatei
	imageFile, err := os.Open(imagePath)
	if err != nil {
		http.Error(w, "Bild nicht gefunden", http.StatusNotFound)
		log.Printf("Fehler beim Öffnen des Bildes: %v", err)
		return
	}
	defer imageFile.Close()

	// Kopieren des Bildes in die HTTP-Antwort
	_, copyErr := io.Copy(w, imageFile)
	if copyErr != nil {
		log.Printf("Fehler beim Senden des Bildes: %v", copyErr)
		return
	}
}
