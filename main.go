package main

//TODO zrobić opcję wpisania reconect czy coś jeżeli bot się rozłączy
//czy inny choi
import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// vars
var public_link string = "http://192.168.0.189"
var test_link string = "http://localhost"

var audio_test int = 3001

// var audio_port int = 81

var main_link string = test_link
var station_name string

type ResponseBody struct {
	Ok int `json:"ok"`
}

func main() {
	fmt.Println("enter the name of the station:")
	fmt.Scanln(&station_name)
	// Wykonaj żądanie GET na /ping
	// response, err := http.Get(main_link + ":3001/connect/" + station_name)
	response, err := http.Get(main_link + ":3001/connect/" + station_name)
	if err != nil {
		fmt.Println("Wystąpił błąd podczas wykonywania żądania:", err)
		return
	}
	defer response.Body.Close()

	// Odczytaj odpowiedź
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Wystąpił błąd podczas odczytywania odpowiedzi:", err)
		return
	}

	// Parsuj odpowiedź JSON
	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		fmt.Println("Wystąpił błąd podczas parsowania JSON:", err)
		return
	}

	// Sprawdź wartość pola "ok"
	if responseBody.Ok == 200 {
		fmt.Println("Odpowiedź z serwera: OK (status 200)")
		start_server()
	} else {
		fmt.Println("Odpowiedź z serwera: Niepoprawny status")
	}
}

func start_server() {
	// Pobierz ścieżkę do katalogu, w którym znajduje się wykonywalny plik
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Wystąpił błąd podczas pobierania ścieżki wykonywalnego pliku: %v", err)
	}
	// Pobierz ścieżkę katalogu nadrzędnego (folder, w którym znajduje się plik wykonywalny)
	baseDirectory := filepath.Dir(executablePath)
	audioDirectory := filepath.Join(baseDirectory, "audio")

	// Definiujemy handler, który zostanie wywołany przy każdym żądaniu na ścieżce "/endpoint"
	http.HandleFunc("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Otrzymano żądanie od serwera")

		// Pobierz listę plików z katalogu /audio
		fileNames, err := listFilesInDirectory(audioDirectory)
		if err != nil {
			log.Fatalf("Wystąpił błąd podczas pobierania listy plików: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Serializuj listę plików do formatu JSON
		responseJSON, err := json.Marshal(fileNames)
		if err != nil {
			log.Fatalf("Wystąpił błąd podczas serializacji listy plików do JSON: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Wyślij listę plików jako odpowiedź
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	http.HandleFunc("/get_song/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Otrzymano żądanie od serwera")

		// Pobieramy nazwę pliku z parametru ścieżki
		fileName := filepath.Base(r.URL.Path)

		// Ścieżka do katalogu /audio
		audioDirectory := "./audio" // Tutaj podaj odpowiednią ścieżkę do katalogu audio

		// Otwieramy plik
		filePath := filepath.Join(audioDirectory, fileName)
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Błąd podczas otwierania pliku: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Kopiujemy zawartość pliku do odpowiedzi HTTP
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("Błąd podczas kopiowania zawartości pliku do odpowiedzi HTTP: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Uruchomienie serwera na porcie 3000
	fmt.Println("Serwer nasłuchuje na porcie 3002...")
	if err := http.ListenAndServe(":3002", nil); err != nil {
		log.Fatalf("Wystąpił błąd podczas uruchamiania serwera: %v", err)
	}
}

func listFilesInDirectory(directoryPath string) ([]string, error) {
	var fileNames []string

	// Odczytaj zawartość katalogu
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	// Iteruj przez pliki i dodaj nazwy do slice'a
	for _, file := range files {
		if file.IsDir() {
			// Pomijaj katalogi, jeśli chcesz tylko pliki
			continue
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}
