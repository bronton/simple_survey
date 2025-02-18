package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Настройка параметров командной строки
	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	port := flag.Int("port", 8080, "Порт для веб-сервера")
	flag.Parse()

	// Загрузка конфигурации
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация хранилища ответов
	responseHandler := NewResponseHandler()

	// Инициализация аудио рекордера
	audioRecorder := NewAudioRecorder()

	// Инициализация обработчика сессий
	sessionManager := NewSessionManager(config, responseHandler, audioRecorder)

	// Настройка HTTP маршрутов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/survey", http.StatusFound)
	})
	http.HandleFunc("/survey", sessionManager.HandleSurveyPage)
	http.HandleFunc("/submit", sessionManager.HandleSubmit)
	http.HandleFunc("/start-recording", sessionManager.HandleStartRecording)
	http.HandleFunc("/stop-recording", sessionManager.HandleStopRecording)
	http.HandleFunc("/complete", sessionManager.HandleComplete)

	// Обработка статических файлов
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запуск сервера
	serverAddr := fmt.Sprintf(":%d", *port)
	server := &http.Server{Addr: serverAddr}

	// Канал для отслеживания сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в горутине
	go func() {
		log.Printf("Сервер запущен на http://localhost%s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stop
	log.Println("Завершение работы сервера...")

	// Очистка ресурсов перед завершением
	sessionManager.Cleanup()
	log.Println("Сервер остановлен")
}
