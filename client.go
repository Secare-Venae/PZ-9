package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Подключение к серверу (измените адрес при необходимости)
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Не удалось подключиться к серверу:", err)
	}
	defer conn.Close()

	// Канал для вывода полученных сообщений (ёмкость 5)
	incoming := make(chan string, 5)

	// Горутина: чтение от сервера → отправка в канал
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			incoming <- scanner.Text()
		}
		close(incoming) // Закрываем канал при разрыве соединения
	}()

	// Горутина: вывод из канала на экран
	go func() {
		for msg := range incoming {
			fmt.Println(msg)
		}
		fmt.Println("\n[СИСТЕМА] Соединение с сервером разорвано")
		os.Exit(0)
	}()

	// Основной цикл: ввод от пользователя
	fmt.Println("Подключено к чату. Введите сообщение:")
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !stdin.Scan() {
			break
		}
		text := strings.TrimSpace(stdin.Text())
		if text == "" {
			continue
		}
		fmt.Fprintln(conn, text)
	}
}