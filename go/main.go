package main

import (
	"fmt"
	_ "html/template"
	"io"
        "log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
        http.HandleFunc("/view/", viewHandler)

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Hier können Sie eine HTML-Templatedatei für die Homepage erstellen
	// und sie mit template.Execute laden.
	fmt.Fprint(w, "Willkommen! Besuchen Sie /upload, um Bilder hochzuladen.")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Fehler beim Lesen der Datei", http.StatusInternalServerError)
			log.Printf("Fehler beim Lesen der Datei: %v", err)
			return
		}
		defer file.Close()

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

                // Erfolgsmeldung mit dem Link ausgeben
		link := fmt.Sprintf("Bild erfolgreich hochgeladen. Sie können es hier anzeigen: <a href='/view/%s'>Anzeigen</a>", handler.Filename)

		// Schreiben Sie den Link als HTML in die Antwort
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, link)
	} else {
		// HTML-Formular für Bild-Upload hier anzeigen
		form := `<html><body>
					<form action="/upload" method="post" enctype="multipart/form-data">
						<input type="file" name="image">
						<input type="submit" value="Hochladen">
					</form>
				</body></html>`
		fmt.Fprint(w, form)
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
