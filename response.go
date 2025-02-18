package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ResponseHandler управляет сохранением и обработкой ответов
type ResponseHandler struct {
	responsesDir string
	mu           sync.Mutex
}

// NewResponseHandler создает новый обработчик ответов
func NewResponseHandler() *ResponseHandler {
	// Создаем директорию для ответов, если она не существует
	responsesDir := "uploads/responses"
	if err := os.MkdirAll(responsesDir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию для ответов: %v", err)
	}

	return &ResponseHandler{
		responsesDir: responsesDir,
		mu:           sync.Mutex{},
	}
}

// SaveResponses сохраняет ответы пользователя в CSV файл
func (rh *ResponseHandler) SaveResponses(sessionID string, responses map[string][]string, 
	questions []QuestionData) error {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	// Формируем имя файла
	fileName := fmt.Sprintf("responses_%s.csv", sessionID)
	filePath := filepath.Join(rh.responsesDir, fileName)

	// Создаем файл
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл %s: %w", filePath, err)
	}
	defer file.Close()

	// Создаем CSV писателя
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовок
	header := []string{"Вопрос ID", "Текст вопроса", "Тип вопроса", "Ответ", "Время ответа"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("ошибка записи заголовка CSV: %w", err)
	}

	// Получаем текущее время
	timestamp := time.Now().Format(time.RFC3339)

	// Записываем ответы
	for _, q := range questions {
		answerValues := responses[q.ID]
		answer := strings.Join(answerValues, "; ")

		record := []string{
			q.ID,
			q.Text,
			string(q.Type),
			answer,
			timestamp,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("ошибка записи ответа в CSV: %w", err)
		}
	}

	return nil
}

// GetResponseFile возвращает путь к файлу с ответами для сессии
func (rh *ResponseHandler) GetResponseFile(sessionID string) (string, error) {
	fileName := fmt.Sprintf("responses_%s.csv", sessionID)
	filePath := filepath.Join(rh.responsesDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл с ответами не найден: %w", err)
	}

	return filePath, nil
}
