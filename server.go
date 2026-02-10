package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

// Client представляет подключенного клиента
type Client struct {
	conn     net.Conn
	username string
}

func main() {
	// Канал для вывода сообщений (ёмкость 10)
	broadcast := make(chan string, 10)

	// Мапа подключенных клиентов с мьютексом для потокобезопасности
	clients := make(map[net.Conn]*Client)
	clientsMutex := &sync.Mutex{}

	// Горутина: вывод сообщений на экран
	go func() {
		for msg := range broadcast {
			fmt.Println(msg)
		}
	}()

	// Горутина: рассылка сообщений всем клиентам
	go func() {
		for msg := range broadcast {
			clientsMutex.Lock()
			for conn := range clients {
				fmt.Fprintln(conn, msg)
			}
			clientsMutex.Unlock()
		}
	}()

	// Запуск сервера (слушает все интерфейсы)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
	defer listener.Close()

	broadcast <- "Сервер запущен на :8080. Ожидание подключений..."

	// Принимаем подключения
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Ошибка принятия подключения:", err)
			continue
		}

		// Генерируем имя клиента
		clientsMutex.Lock()
		username := fmt.Sprintf("Пользователь_%d", len(clients)+1)
		clients[conn] = &Client{conn: conn, username: username}
		clientsMutex.Unlock()

		broadcast <- fmt.Sprintf("[СИСТЕМА] %s подключился (%s)", username, conn.RemoteAddr())

		// Обработка клиента в отдельной горутине
		go handleClient(conn, username, broadcast, clients, clientsMutex)
	}
}

func handleClient(conn net.Conn, username string, broadcast chan<- string, clients map[net.Conn]*Client, mutex *sync.Mutex) {
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
		broadcast <- fmt.Sprintf("[СИСТЕМА] %s отключился", username)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		if msg == "" {
			continue
		}
		// Отправляем сообщение всем клиентам через канал
		broadcast <- fmt.Sprintf("[%s] %s", username, msg)
	}

	if err := scanner.Err(); err != nil {
		broadcast <- fmt.Sprintf("[СИСТЕМА] Ошибка чтения от %s: %v", username, err)
	}
}