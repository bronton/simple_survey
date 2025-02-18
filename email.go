package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"github.com/jordan-wright/email"
)

// Emailer отправляет результаты на email
type Emailer struct {
	config *Config
}

// NewEmailer создает новый экземпляр отправителя email
func NewEmailer(config *Config) *Emailer {
	return &Emailer{
		config: config,
	}
}

// SendZipResults отправляет zip-архив с результатами на email
func (e *Emailer) SendZipResults(zipPath, sessionID string) error {
	// Проверяем существование архива
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return fmt.Errorf("архив не найден: %w", err)
	}

	// Создаем новое email сообщение
	em := email.NewEmail()
	em.From = e.config.Email.From
	em.To = []string{e.config.Email.To}
	
	// Формируем тему письма
	subject := e.config.Email.Subject
	if subject == "" {
		subject = "Результаты опроса"
	}
	em.Subject = fmt.Sprintf("%s - Сессия %s - %s", 
		subject, 
		sessionID[:8], // Используем первые 8 символов ID для краткости
		time.Now().Format("2006-01-02 15:04"))

	// Тело письма
	em.Text = []byte(fmt.Sprintf(`Здравствуйте!

Во вложении находятся результаты опроса, проведенного %s.

ID сессии: %s
Время завершения: %s

В архиве содержатся:
1. CSV-файл с ответами пользователя
2. Аудиозапись, сделанная во время прохождения опроса

С уважением,
Система автоматического тестирования
`, 
		time.Now().Format("02.01.2006 в 15:04"),
		sessionID,
		time.Now().Format("02.01.2006 15:04:05")))

	// Прикрепляем файл архива
	if _, err := em.AttachFile(zipPath); err != nil {
		return fmt.Errorf("ошибка прикрепления файла: %w", err)
	}

	// Формируем адрес SMTP сервера с аутентификацией
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)
	auth := smtp.PlainAuth("", e.config.SMTPUser, e.config.SMTPPass, e.config.SMTPHost)

	// Отправляем email
	if err := em.Send(addr, auth); err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}

	log.Printf("Email с результатами успешно отправлен на %s", e.config.Email.To)
	return nil
}
