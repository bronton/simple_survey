package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session представляет сессию тестирования пользователя
type Session struct {
	ID            string
	StartTime     time.Time
	AudioFilePath string
	Completed     bool
	Responses     map[string][]string
}

// SessionManager управляет сессиями пользователей
type SessionManager struct {
	config         *Config
	sessions       map[string]*Session
	responseHandler *ResponseHandler
	audioRecorder   *AudioRecorder
	templates      *template.Template
	mu             sync.RWMutex
}

// NewSessionManager создает новый менеджер сессий
func NewSessionManager(config *Config, responseHandler *ResponseHandler, audioRecorder *AudioRecorder) *SessionManager {
	// Загрузка шаблонов
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Ошибка загрузки шаблонов: %v", err)
	}

	return &SessionManager{
		config:         config,
		sessions:       make(map[string]*Session),
		responseHandler: responseHandler,
		audioRecorder:   audioRecorder,
		templates:      tmpl,
		mu:             sync.RWMutex{},
	}
}

// newSession создает новую сессию
func (sm *SessionManager) newSession() *Session {
	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		StartTime: time.Now(),
		Responses: make(map[string][]string),
	}
	
	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()
	
	return session
}

// getSession получает сессию по ID
func (sm *SessionManager) getSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[id]
	return session, exists
}

// HandleSurveyPage обрабатывает запрос на страницу с опросом
func (sm *SessionManager) HandleSurveyPage(w http.ResponseWriter, r *http.Request) {
	// Создаем новую сессию для каждого посещения страницы опроса
	session := sm.newSession()
	
	// Устанавливаем cookie с ID сессии
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, // 1 час
	}
	http.SetCookie(w, &cookie)
	
	// Отображаем шаблон с вопросами
	data := struct {
		Questions []QuestionData
		SessionID string
	}{
		Questions: sm.config.Questions,
		SessionID: session.ID,
	}
	
	if err := sm.templates.ExecuteTemplate(w, "survey.html", data); err != nil {
		log.Printf("Ошибка отображения шаблона survey.html: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

// HandleStartRecording начинает запись аудио
func (sm *SessionManager) HandleStartRecording(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Отсутствует ID сессии", http.StatusBadRequest)
		return
	}
	
	session, exists := sm.getSession(sessionID)
	if !exists {
		http.Error(w, "Недействительная сессия", http.StatusBadRequest)
		return
	}
	
	// Создаем путь для сохранения аудио
	audioPath := filepath.Join("uploads", fmt.Sprintf("audio_%s.wav", sessionID))
	
	// Начинаем запись
	if err := sm.audioRecorder.StartRecording(sessionID, audioPath); err != nil {
		log.Printf("Ошибка начала записи: %v", err)
		http.Error(w, "Не удалось начать запись", http.StatusInternalServerError)
		return
	}
	
	session.AudioFilePath = audioPath
	w.WriteHeader(http.StatusOK)
}

// HandleStopRecording останавливает запись аудио
func (sm *SessionManager) HandleStopRecording(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Отсутствует ID сессии", http.StatusBadRequest)
		return
	}
	
	// Останавливаем запись
	if err := sm.audioRecorder.StopRecording(sessionID); err != nil {
		log.Printf("Ошибка остановки записи: %v", err)
		http.Error(w, "Не удалось остановить запись", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// HandleSubmit обрабатывает отправку формы с ответами
func (sm *SessionManager) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}
	
	sessionID := r.FormValue("session_id")
	if sessionID == "" {
		http.Error(w, "Отсутствует ID сессии", http.StatusBadRequest)
		return
	}
	
	session, exists := sm.getSession(sessionID)
	if !exists {
		http.Error(w, "Недействительная сессия", http.StatusBadRequest)
		return
	}
	
	// Сохраняем ответы
	for _, question := range sm.config.Questions {
		values := r.Form[question.ID]
		// Для вопросов с произвольным ответом добавляем его отдельно
		if question.AllowCustom {
			customAnswer := r.FormValue(question.ID + "_custom")
			if customAnswer != "" {
				values = append(values, customAnswer)
			}
		}
		session.Responses[question.ID] = values
	}
	
	// Останавливаем запись аудио, если она не была остановлена ранее
	sm.audioRecorder.StopRecording(sessionID)
	
	// Сохраняем ответы
	if err := sm.responseHandler.SaveResponses(session.ID, session.Responses, sm.config.Questions); err != nil {
		log.Printf("Ошибка сохранения ответов: %v", err)
		http.Error(w, "Не удалось сохранить ответы", http.StatusInternalServerError)
		return
	}
	
	// Отправляем результаты на email
	if err := sm.SendResults(session); err != nil {
		log.Printf("Ошибка отправки результатов: %v", err)
		// Продолжаем выполнение, даже если отправка не удалась
	}
	
	session.Completed = true
	
	// Перенаправляем на страницу завершения
	http.Redirect(w, r, "/complete?session_id="+sessionID, http.StatusSeeOther)
}

// HandleComplete отображает страницу завершения
func (sm *SessionManager) HandleComplete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Отсутствует ID сессии", http.StatusBadRequest)
		return
	}
	
	session, exists := sm.getSession(sessionID)
	if !exists || !session.Completed {
		http.Redirect(w, r, "/survey", http.StatusSeeOther)
		return
	}
	
	if err := sm.templates.ExecuteTemplate(w, "complete.html", nil); err != nil {
		log.Printf("Ошибка отображения шаблона complete.html: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

// SendResults отправляет результаты на email
func (sm *SessionManager) SendResults(session *Session) error {
	// Получаем путь к CSV файлу с ответами
	csvPath, err := sm.responseHandler.GetResponseFile(session.ID)
	if err != nil {
		return fmt.Errorf("не удалось получить файл с ответами: %w", err)
	}
	
	// Создаем архив
	zipPath := filepath.Join("uploads", fmt.Sprintf("results_%s.zip", session.ID))
	files := []string{csvPath}
	
	// Добавляем аудио файл, если он существует
	if session.AudioFilePath != "" {
		files = append(files, session.AudioFilePath)
	}
	
	if err := CreateZipArchive(zipPath, files); err != nil {
		return fmt.Errorf("ошибка создания архива: %w", err)
	}
	
	// Отправляем архив по email
	emailer := NewEmailer(sm.config)
	if err := emailer.SendZipResults(zipPath, session.ID); err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}
	
	return nil
}

// Cleanup удаляет временные файлы и ресурсы
func (sm *SessionManager) Cleanup() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Останавливаем все активные записи
	for id := range sm.sessions {
		sm.audioRecorder.StopRecording(id)
	}
}
