package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
    // Определение флагов командной строки
    webdavURL := flag.String("webdav-url", "", "URL WebDAV сервера")
    remoteFolder := flag.String("remote-folder", "", "Путь к каталогу на сервере")
    username := flag.String("username", "", "Имя пользователя для WebDAV")
    password := flag.String("password", "", "Пароль для WebDAV")
    localFolder := flag.String("local-folder", "", "Локальная папка для синхронизации")
    logFileName := flag.String("log-file", "sync_log.txt", "Имя файла лога")
    fileMask := flag.String("file-mask", "*", "Маска для фильтрации файлов (например, *.txt или *report*)")
    
    // Парсим флаги командной строки
    flag.Parse()

    // Проверка обязательных параметров
    if *webdavURL == "" || *remoteFolder == "" || *username == "" || *password == "" || *localFolder == "" {
        log.Fatal("Ошибка: все параметры, кроме лог-файла и маски, обязательны.")
    }

    // Настроить логирование
    logFilePath := filepath.Join(*localFolder, *logFileName)
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Ошибка открытия файла лога: %v", err)
    }
    defer logFile.Close()

    // Настроить лог на запись в файл и в консоль
    log.SetOutput(io.MultiWriter(os.Stdout, logFile))
    log.Println("Запуск синхронизации:", time.Now().Format(time.RFC1123))

    // Получаем список файлов на WebDAV сервере
    remoteFiles, err := getRemoteFiles(*webdavURL, *remoteFolder, *username, *password)
    if err != nil {
        log.Fatalf("Ошибка получения списка файлов с сервера: %v", err)
    }

    // Получаем список файлов в локальной директории
    localFiles, err := getLocalFiles(*localFolder, *fileMask)
    if err != nil {
        log.Fatalf("Ошибка получения списка файлов из локального каталога: %v", err)
    }

    // Сравниваем и загружаем файлы, отсутствующие на сервере
    for _, localFile := range localFiles {
        // Получаем только имя файла
        fileName := filepath.Base(localFile)

        // Если файла нет на сервере, загружаем его
        if _, exists := remoteFiles[fileName]; !exists {
            log.Printf("Файл %s отсутствует на сервере. Загружаем...\n", fileName)
            if err := uploadFile(localFile, fileName, *webdavURL, *remoteFolder, *username, *password); err != nil {
                log.Printf("Ошибка загрузки файла %s: %v", fileName, err)
            } else {
                log.Printf("Файл %s успешно загружен.\n", fileName)
            }
        }
    }

    log.Println("Синхронизация завершена:", time.Now().Format(time.RFC1123))
}

// Структуры для парсинга XML ответа WebDAV

type MultiStatus struct {
    XMLName    xml.Name     `xml:"multistatus"`
    Responses  []Response   `xml:"response"`
}

type Response struct {
    Href    string `xml:"href"`
    Propstat Propstat `xml:"propstat"`
}

type Propstat struct {
    Status string `xml:"status"`
}

// getRemoteFiles получает список файлов в указанном каталоге на WebDAV сервере
func getRemoteFiles(webdavURL, remoteFolder, username, password string) (map[string]struct{}, error) {
    files := make(map[string]struct{})

    // Создаем полный URL к целевому каталогу на сервере
    fullURL := webdavURL + remoteFolder

    // Создаем запрос с методом PROPFIND
    req, err := http.NewRequest("PROPFIND", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("ошибка создания запроса: %w", err)
    }
    req.SetBasicAuth(username, password)
    req.Header.Set("Depth", "1")
    req.Header.Set("Content-Type", "text/xml")

    // Выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
    }
    defer resp.Body.Close()

    // Чтение тела ответа
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
    }

    // Парсим XML ответ
    var multistatus MultiStatus
    if err := xml.Unmarshal(body, &multistatus); err != nil {
        return nil, fmt.Errorf("ошибка разбора XML: %w", err)
    }

    // Добавляем каждый файл на сервере в карту
    for _, response := range multistatus.Responses {
        decodedHref, err := url.PathUnescape(strings.TrimSpace(response.Href))
        if err != nil {
            log.Printf("Ошибка декодирования %s: %v", response.Href, err)
            continue
        }
        // Получаем только имя файла
        fileName := filepath.Base(decodedHref)
        files[fileName] = struct{}{}
    }

    return files, nil
}

// getLocalFiles получает список файлов из локального каталога с фильтрацией по маске
func getLocalFiles(folder, mask string) ([]string, error) {
    var files []string

    // Создаем регулярное выражение для фильтрации
    regex, err := regexp.Compile(mask)
    if err != nil {
        return nil, fmt.Errorf("ошибка компиляции регулярного выражения: %w", err)
    }

    err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && regex.MatchString(info.Name()) {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("ошибка чтения локального каталога: %w", err)
    }
    return files, nil
}

// uploadFile загружает файл на WebDAV сервер в указанный каталог
func uploadFile(localPath, fileName, webdavURL, remoteFolder, username, password string) error {
    // Формируем полный путь для загрузки файла на сервере
    remotePath := webdavURL + remoteFolder + url.PathEscape(fileName)

    // Открываем файл для чтения
    file, err := os.Open(localPath)
    if err != nil {
        return fmt.Errorf("ошибка открытия файла %s: %w", localPath, err)
    }
    defer file.Close()

    // Создаем запрос PUT для загрузки файла
    req, err := http.NewRequest("PUT", remotePath, file)
    if err != nil {
        return fmt.Errorf("ошибка создания запроса PUT для %s: %w", fileName, err)
    }
    req.SetBasicAuth(username, password)

    // Выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("ошибка выполнения запроса PUT для %s: %w", fileName, err)
    }
    defer resp.Body.Close()

    // Проверяем статус ответа
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("не удалось загрузить файл %s, статус: %s", fileName, resp.Status)
    }

    return nil
}
