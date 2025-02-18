package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateZipArchive создает zip-архив с указанными файлами
func CreateZipArchive(zipPath string, filePaths []string) error {
	// Создаем директорию для архива, если нужно
	dir := filepath.Dir(zipPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", dir, err)
	}

	// Создаем файл архива
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл архива: %w", err)
	}
	defer zipFile.Close()

	// Создаем zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Добавляем файлы в архив
	for _, filePath := range filePaths {
		if err := addFileToZip(zipWriter, filePath); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip добавляет файл в zip архив
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	// Проверяем существование файла
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о файле %s: %w", filePath, err)
	}

	// Проверяем, что это файл, а не директория
	if fileInfo.IsDir() {
		return fmt.Errorf("%s является директорией, а не файлом", filePath)
	}

	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", filePath, err)
	}
	defer file.Close()

	// Создаем файл внутри архива
	baseFileName := filepath.Base(filePath)
	zipFile, err := zipWriter.Create(baseFileName)
	if err != nil {
		return fmt.Errorf("не удалось создать файл в архиве: %w", err)
	}

	// Копируем содержимое файла в архив
	if _, err := io.Copy(zipFile, file); err != nil {
		return fmt.Errorf("ошибка копирования файла в архив: %w", err)
	}

	return nil
}
