package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	Email     EmailConfig    `json:"email"`
	Questions []QuestionData `json:"questions"`
	SMTPHost  string         `json:"smtp_host"`
	SMTPPort  int            `json:"smtp_port"`
	SMTPUser  string         `json:"smtp_user"`
	SMTPPass  string         `json:"smtp_pass"`
}

// EmailConfig содержит настройки получателя email
type EmailConfig struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
}

// QuestionType определяет тип вопроса
type QuestionType string

const (
	TypeSingleChoice QuestionType = "single_choice"
	TypeMultiChoice  QuestionType = "multi_choice"
	TypeText         QuestionType = "text"
	TypeMixed        QuestionType = "mixed"
)

// QuestionData представляет структуру вопроса
type QuestionData struct {
	ID          string       `json:"id"`
	Text        string       `json:"text"`
	Type        QuestionType `json:"type"`
	Options     []string     `json:"options,omitempty"`
	AllowCustom bool         `json:"allow_custom,omitempty"`
	Required    bool         `json:"required"`
}

// LoadConfig загружает конфигурацию из JSON файла
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("невозможно открыть файл конфигурации: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig проверяет корректность конфигурации
func validateConfig(config *Config) error {
	if config.Email.To == "" {
		return fmt.Errorf("email получателя не указан")
	}

	if config.Email.From == "" {
		return fmt.Errorf("email отправителя не указан")
	}

	if len(config.Questions) == 0 {
		return fmt.Errorf("список вопросов пуст")
	}

	for i, q := range config.Questions {
		if q.ID == "" {
			return fmt.Errorf("вопрос #%d: отсутствует ID", i+1)
		}
		if q.Text == "" {
			return fmt.Errorf("вопрос #%d: отсутствует текст вопроса", i+1)
		}

		switch q.Type {
		case TypeSingleChoice, TypeMultiChoice:
			if len(q.Options) == 0 {
				return fmt.Errorf("вопрос #%d: тип %s требует наличия вариантов ответа", i+1, q.Type)
			}
		case TypeMixed:
			if len(q.Options) == 0 || !q.AllowCustom {
				return fmt.Errorf("вопрос #%d: тип mixed требует наличия вариантов ответа и разрешения ввода пользователя", i+1)
			}
		case TypeText:
			// Для текстовых вопросов нет специальных требований
		default:
			return fmt.Errorf("вопрос #%d: неизвестный тип %s", i+1, q.Type)
		}
	}

	// Проверка SMTP настроек
	if config.SMTPHost == "" || config.SMTPPort == 0 {
		return fmt.Errorf("неверные настройки SMTP сервера")
	}

	return nil
}
