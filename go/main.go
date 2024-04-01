package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
        "crypto/rand"
        "encoding/base64"
)

var (
	lastUploadTime time.Time
	mu             sync.Mutex
	uploadInterval = 10 * time.Second
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/image/", imageHandler)
	http.HandleFunc("/view/", viewHandler)
        // Statischen Dateipfad setzen
        fs := http.FileServer(http.Dir("static"))
        http.Handle("/static/", http.StripPrefix("/static/", fs))

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

func generateNonce() (string, error) {
    nonceBytes := make([]byte, 16) // 16 Bytes generieren eine ausreichend lange Zeichenfolge für den Nonce
    if _, err := rand.Read(nonceBytes); err != nil {
        return "", err // Im Fehlerfall, geben Sie den Fehler zurück
    }
    return base64.StdEncoding.EncodeToString(nonceBytes), nil
}


func uploadHandler(w http.ResponseWriter, r *http.Request) {
        nonce, err := generateNonce()
    if err != nil {
        // Fehlerbehandlung, z.B. Senden eines Serverfehlers
        http.Error(w, "Serverfehler", http.StatusInternalServerError)
        log.Printf("Fehler beim Generieren des Nonce: %v", err)
        return
    }

	// Setzen der Content Security Policy
        //w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")
        //w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self'; script-src 'self' 'nonce-%s'; object-src 'none';", nonce))
        w.Header().Set("Content-Security-Policy", fmt.Sprintf("script-src 'self' 'nonce-%s';", nonce))

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
		if !strings.HasPrefix(mimeType, "image/") && !strings.HasPrefix(mimeType, "text/xml") && !strings.HasPrefix(mimeType, "image/svg+xml") {
			http.Error(w, "Nur Bild-Uploads sind erlaubt", http.StatusBadRequest)
			log.Printf("Versuch, eine Nicht-Bild-Datei hochzuladen: %v", mimeType)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, "Fehler beim Zurücksetzen des Dateizeigers", http.StatusInternalServerError)
			log.Printf("Fehler beim Zurücksetzen des Dateizeigers: %v", err)
			return
		}

		// Ermitteln, ob der ursprüngliche Dateiname erzwungen werden soll
		//forceName := r.Header.Get("Force-Name")
		forceName := r.FormValue("force_name")
		var filename string
		if forceName == "true" {
			filename = handler.Filename
		} else {
			// Zeitstempel zum Dateinamen hinzufügen
			timestamp := time.Now().Format("20060102-150405")
			filename = fmt.Sprintf("%s-%s", timestamp, handler.Filename)
		}

		// Datei speichern
		uploadPath := "./uploads/" + filename
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

		tmpl, err := template.ParseFiles("templates/uploadSuccess.html")
		if err != nil {
			http.Error(w, "Fehler beim Laden des Templates", http.StatusInternalServerError)
			log.Printf("Fehler beim Laden des Templates: %v", err)
			return
		}

		data := struct {
			Message  string
			Filename string
                        Nonce    string
		}{
			Message:  "Bild erfolgreich hochgeladen.",
			Filename: filename, // Geändert, um den möglicherweise modifizierten Dateinamen anzuzeigen
                        Nonce:    nonce,
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

// Funktion zur Ermittlung des MIME-Types basierend auf der Dateiendung
func getMimeType(filePath string) string {
	switch filepath.Ext(filePath) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	// Setzen der Content Security Policy
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")

	// Extrahieren des Bildnamens aus dem URL-Pfad
	imagePath := r.URL.Path[len("/image/"):]

	// Reinigen des Pfades, um Directory Traversal zu verhindern
	cleanedPath := path.Clean("/uploads/" + imagePath)

	// Generieren des absoluten Pfads zum uploads-Verzeichnis
	uploadsDir, err := filepath.Abs("./uploads")
	if err != nil {
		http.Error(w, "Interner Serverfehler", http.StatusInternalServerError)
		log.Printf("Fehler beim Ermitteln des absoluten Pfads des uploads-Verzeichnisses: %v", err)
		return
	}

	// Generieren des absoluten Pfads zur angeforderten Datei
	absImagePath, err := filepath.Abs(cleanedPath)
	if err != nil {
		http.Error(w, "Interner Serverfehler", http.StatusInternalServerError)
		log.Printf("Fehler beim Ermitteln des absoluten Pfads des Bildes: %v", err)
		return
	}

	// Sicherstellen, dass das Bild im uploads-Verzeichnis liegt
	if !strings.HasPrefix(absImagePath, uploadsDir) {
		http.Error(w, "Zugriff verweigert", http.StatusForbidden)
		log.Printf("Versuch, auf Datei außerhalb des uploads-Verzeichnisses zuzugreifen: %v", absImagePath)
		return
	}

	// Stellen Sie sicher, dass das Bild existiert
	if _, err := os.Stat(absImagePath); os.IsNotExist(err) {
		http.Error(w, "Bild nicht gefunden", http.StatusNotFound)
		log.Printf("Bild nicht gefunden: %v", err)
		return
	}

	// Setzen der korrekten MIME-Type basierend auf der Dateiendung
	mimeType := getMimeType(imagePath)
	w.Header().Set("Content-Type", mimeType)

	// Ausliefern des Bildes
	http.ServeFile(w, r, absImagePath)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// Setzen der Content Security Policy
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")

	filePath := r.URL.Path[len("/view/"):]
	imagePath := "./uploads/" + filePath

	// Überprüfen, ob die Bilddatei existiert
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Bild nicht gefunden", http.StatusNotFound)
		log.Printf("Bild nicht gefunden: %v", err)
		return
	}

	// Verwenden von html/template zur sicheren Ausgabe von HTML
	tmpl, err := template.ParseFiles("templates/viewImage.html")
	if err != nil {
		http.Error(w, "Fehler beim Laden des Templates", http.StatusInternalServerError)
		log.Printf("Fehler beim Laden des Templates: %v", err)
		return
	}

	data := struct {
		Filename string
	}{
		Filename: filePath,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Fehler beim Rendern des Templates", http.StatusInternalServerError)
		log.Printf("Fehler beim Rendern des Templates: %v", err)
	}
}
