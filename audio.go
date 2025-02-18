package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

const (
	sampleRate   = 44100
	numChannels  = 1
	bitsPerSample = 16
)

// AudioRecorder управляет записью аудио
type AudioRecorder struct {
	recordings map[string]*Recording
	mu         sync.Mutex
}

// Recording представляет активную запись аудио
type Recording struct {
	stream     *portaudio.Stream
	buffer     []int16
	bufferLock sync.Mutex
	filePath   string
	stopChan   chan struct{}
}

// NewAudioRecorder создает новый аудио рекордер
func NewAudioRecorder() *AudioRecorder {
	// Инициализация portaudio
	if err := portaudio.Initialize(); err != nil {
		log.Fatalf("Ошибка инициализации portaudio: %v", err)
	}

	// Создаем директорию для загрузок, если она не существует
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("Не удалось создать директорию uploads: %v", err)
	}

	return &AudioRecorder{
		recordings: make(map[string]*Recording),
		mu:         sync.Mutex{},
	}
}

// StartRecording начинает запись аудио для сессии
func (ar *AudioRecorder) StartRecording(sessionID, filePath string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Проверяем, существует ли уже запись для этой сессии
	if _, exists := ar.recordings[sessionID]; exists {
		return fmt.Errorf("запись для сессии %s уже запущена", sessionID)
	}

	// Создаем директорию для файла, если нужно
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", dir, err)
	}

	// Инициализируем запись
	recording := &Recording{
		buffer:   make([]int16, 0),
		filePath: filePath,
		stopChan: make(chan struct{}),
	}

	// Открываем поток аудио
	inputChannels := numChannels
	outputChannels := 0 // Нам не нужен выходной канал
	framesPerBuffer := 1024

	var err error
	recording.stream, err = portaudio.OpenDefaultStream(
		inputChannels, outputChannels, float64(sampleRate),
		framesPerBuffer, recording.processAudio,
	)
	if err != nil {
		return fmt.Errorf("не удалось открыть аудио поток: %w", err)
	}

	// Запускаем поток
	if err := recording.stream.Start(); err != nil {
		recording.stream.Close()
		return fmt.Errorf("не удалось запустить аудио поток: %w", err)
	}

	// Сохраняем запись
	ar.recordings[sessionID] = recording

	// Запускаем горутину для обработки запроса на остановку
	go func() {
		<-recording.stopChan
		recording.stream.Stop()
		recording.stream.Close()

		// Сохраняем файл WAV
		if err := ar.saveWavFile(recording); err != nil {
			log.Printf("Ошибка сохранения WAV файла: %v", err)
		}
	}()

	return nil
}

// processAudio обрабатывает входящие аудио данные
func (r *Recording) processAudio(in []int16) {
	r.bufferLock.Lock()
	defer r.bufferLock.Unlock()
	
	// Копируем данные в буфер
	r.buffer = append(r.buffer, in...)
}

// StopRecording останавливает запись аудио
func (ar *AudioRecorder) StopRecording(sessionID string) error {
	ar.mu.Lock()
	recording, exists := ar.recordings[sessionID]
	if !exists {
		ar.mu.Unlock()
		return nil // Запись уже остановлена или не существовала
	}
	
	// Удаляем запись из мапы
	delete(ar.recordings, sessionID)
	ar.mu.Unlock()
	
	// Отправляем сигнал остановки
	close(recording.stopChan)
	
	return nil
}

// saveWavFile сохраняет буфер аудио в WAV файл
func (ar *AudioRecorder) saveWavFile(recording *Recording) error {
	recording.bufferLock.Lock()
	defer recording.bufferLock.Unlock()
	
	// Открываем файл для записи
	file, err := os.Create(recording.filePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл %s: %w", recording.filePath, err)
	}
	defer file.Close()
	
	// Создаем WAV писателя
	writer := wav.NewWriter(file, uint32(len(recording.buffer)),
		uint16(numChannels), uint32(sampleRate), uint16(bitsPerSample))
	
	// Преобразуем int16 данные в байты
	samples := make([]wav.Sample, len(recording.buffer))
	for i, s := range recording.buffer {
		samples[i] = wav.Sample{Values: []int{int(s)}}
	}
	
	// Записываем данные
	if err := writer.WriteSamples(samples); err != nil {
		return fmt.Errorf("ошибка записи WAV семплов: %w", err)
	}
	
	return nil
}

// Cleanup освобождает ресурсы
func (ar *AudioRecorder) Cleanup() {
	// Останавливаем все активные записи
	ar.mu.Lock()
	sessions := make([]string, 0, len(ar.recordings))
	for id := range ar.recordings {
		sessions = append(sessions, id)
	}
	ar.mu.Unlock()
	
	for _, id := range sessions {
		ar.StopRecording(id)
	}
	
	// Завершаем работу portaudio
	if err := portaudio.Terminate(); err != nil {
		log.Printf("Ошибка при завершении работы portaudio: %v", err)
	}
}
