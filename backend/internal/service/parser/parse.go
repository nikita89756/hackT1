package parser

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"mime/multipart"
// 	"os"
// 	"path/filepath"

// 	"net/http"
// )

// type ParseResponse struct {
// 	Content string `json:"content"`
// }

// func ConvertPDFToText(pdfPath string) (string, error) {
// 	// Открываем PDF файл
// 	file, err := os.Open(pdfPath)
// 	if err != nil {
// 		return "", fmt.Errorf("error opening file: %v", err)
// 	}
// 	defer file.Close()

// 	// Создаем буфер для multipart form
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)

// 	// Создаем form-file
// 	part, err := writer.CreateFormFile("file", filepath.Base(pdfPath))
// 	if err != nil {
// 		return "", fmt.Errorf("error creating form file: %v", err)
// 	}

// 	// Копируем содержимое файла в form-file
// 	_, err = io.Copy(part, file)
// 	if err != nil {
// 		return "", fmt.Errorf("error copying file: %v", err)
// 	}

// 	err = writer.Close()
// 	if err != nil {
// 		return "", fmt.Errorf("error closing writer: %v", err)
// 	}

// 	// Создаем HTTP запрос
// 	req, err := http.NewRequest("POST", "http://localhost:8000/parse_pdf", body)
// 	if err != nil {
// 		return "", fmt.Errorf("error creating request: %v", err)
// 	}

// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Отправляем запрос
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Читаем ответ
// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading response: %v", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("error response: %s", string(respBody))
// 	}

// 	// Для отладки
// 	fmt.Printf("Response body: %s\n", string(respBody))

// 	// Парсим JSON ответ
// 	var parseResp ParseResponse
// 	if err := json.Unmarshal(respBody, &parseResp); err != nil {
// 		return "", fmt.Errorf("error parsing response: %v\nResponse body: %s", err, string(respBody))
// 	}

// 	return parseResp.Content, nil
// }

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// Конфигурация HTTP клиента
	defaultTimeout     = 30 * time.Second
	handshakeTimeout   = 10 * time.Second
	keepAliveTimeout   = 600 * time.Second
	maxIdleConnections = 100


	// User-Agent для HTTP запросов
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

// ParseResponse представляет ответ от API парсера
type ParseResponse struct {
	Content    string `json:"content"`
	OutputPath string `json:"output_path,omitempty"`
}

// HTTPClient представляет HTTP клиент с настроенными таймаутами
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient создает новый HTTP клиент с оптимальными настройками
func NewHTTPClient(defaultAPIEndpoint string) *HTTPClient {
	transport := &http.Transport{
		TLSHandshakeTimeout:   handshakeTimeout,
		ResponseHeaderTimeout: handshakeTimeout,
		MaxIdleConns:          maxIdleConnections,
		IdleConnTimeout:       keepAliveTimeout,
		DisableKeepAlives:     false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}

	return &HTTPClient{
		client:  client,
		baseURL: defaultAPIEndpoint,
	}
}

// ReadFileContent читает содержимое файла и возвращает его в виде строки
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	return string(content), nil
}

// convertFileToText отправляет файл на API для конвертации
func (c *HTTPClient) convertFileToText(ctx context.Context, filePath, endpoint string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("error copying file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, body)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response (status %d): %s", resp.StatusCode, string(respBody))
	}

	var parseResp ParseResponse
	if err := json.Unmarshal(respBody, &parseResp); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	return parseResp.Content, nil
}

// convertPDFToText конвертирует PDF файл в текст
func (c *HTTPClient) convertPDFToText(ctx context.Context, pdfPath string) (string, error) {
	return c.convertFileToText(ctx, pdfPath, "/parse_pdf")
}

// convertDOCXToText конвертирует DOCX файл в текст
func (c *HTTPClient) convertDOCXToText(ctx context.Context, docxPath string) (string, error) {
	return c.convertFileToText(ctx, docxPath, "/parse_docx")
}

// cleanText очищает текст от лишних пробелов и переносов строк
func cleanText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// findMainContent ищет основной контент на веб-странице
func findMainContent(node *html.Node) *html.Node {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "class" || attr.Key == "id" {
				value := strings.ToLower(attr.Val)
				if strings.Contains(value, "main") ||
					strings.Contains(value, "content") ||
					strings.Contains(value, "article") ||
					strings.Contains(value, "post") ||
					strings.Contains(value, "body-content") {
					return node
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if result := findMainContent(c); result != nil {
			return result
		}
	}

	return nil
}

// extractText извлекает текст из HTML узла
func extractText(node *html.Node) string {
	var buf bytes.Buffer

	// Игнорируемые теги
	ignoreTags := map[string]bool{
		"script": true, "style": true, "noscript": true,
		"iframe": true, "nav": true, "header": true,
		"footer": true, "aside": true, "menu": true,
		"form": true, "button": true,
	}

	// Игнорируемые классы
	ignoreClasses := []string{
		"nav", "navigation", "menu", "header", "footer",
		"sidebar", "widget", "cookie", "popup", "modal",
		"banner", "ad-",
	}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if ignoreTags[n.Data] {
				return
			}

			for _, attr := range n.Attr {
				if attr.Key == "class" || attr.Key == "id" {
					value := strings.ToLower(attr.Val)
					for _, ignore := range ignoreClasses {
						if strings.Contains(value, ignore) {
							return
						}
					}
				}
			}

			if n.Data == "p" || n.Data == "div" || n.Data == "h1" || n.Data == "h2" || n.Data == "h3" {
				buf.WriteString("\n")
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" && len(strings.Fields(text)) > 2 {
				buf.WriteString(text + "\n")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}

		if n.Type == html.ElementNode {
			if n.Data == "p" || n.Data == "div" || n.Data == "h1" || n.Data == "h2" || n.Data == "h3" {
				buf.WriteString("\n")
			}
		}
	}

	extract(node)
	return cleanText(buf.String())
}

// parseURL извлекает текст из веб-страницы
func (c *HTTPClient) parseURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response from URL: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	// Ищем основной контент
	if mainContent := findMainContent(doc); mainContent != nil {
		return extractText(mainContent), nil
	}

	// Ищем body, если основной контент не найден
	var body *html.Node
	var findBody func(*html.Node) bool
	findBody = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if findBody(c) {
				return true
			}
		}
		return false
	}
	findBody(doc)

	if body != nil {
		return extractText(body), nil
	}

	return extractText(doc), nil
}

// ExtractText извлекает текст из файла или URL
func ExtractText(inputFile string,defaultUrl string) (string, error) {
	ctx := context.Background()
	client := NewHTTPClient(defaultUrl)

	switch {
	case strings.HasPrefix(inputFile, "http://") || strings.HasPrefix(inputFile, "https://"):
		return client.parseURL(ctx, inputFile)

	case strings.HasSuffix(inputFile, ".pdf"):
		return client.convertPDFToText(ctx, inputFile)

	case strings.HasSuffix(inputFile, ".docx"):
		return client.convertDOCXToText(ctx, inputFile)

	case strings.HasSuffix(inputFile, ".txt"):
		return ReadFileContent(inputFile)

	default:
		return "", fmt.Errorf("unsupported file type: %s", filepath.Ext(inputFile))
	}
}
