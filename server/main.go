package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Currency struct {
	USDBRL struct {
		Code       string `json:"-"`
		Codein     string `json:"-"`
		Name       string `json:"-"`
		High       string `json:"-"`
		Low        string `json:"-"`
		VarBid     string `json:"-"`
		PctChange  string `json:"-"`
		Bid        string `json:"bid"`
		Ask        string `json:"-"`
		Timestamp  string `json:"-"`
		CreateDate string `json:"-"`
	} `json:"USDBRL"`
}

type BidData struct {
	Bid  string
	Date time.Time
}

func main() {
	http.HandleFunc("/cotacao", handler)
	fmt.Println("servidor rodando na porta 8080....")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("erro ao iniciar o servidor!")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, "erro ao criar requisição", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "erro ao fazer requisição à API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "erro ao ler resposta da API", http.StatusInternalServerError)
		return
	}

	var respBody Currency
	json.Unmarshal(body, &respBody)

	err = connectAndSave(respBody.USDBRL.Bid)
	if err != nil {
		http.Error(w, "erro ao gravar as informações do banco de dados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respBody)
}

func connectAndSave(data string) error {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/goexpert?parseTime=true&loc=America%2FSao_Paulo")
	if err != nil {
		fmt.Println("erro ao conectar ao banco de dados: ", err)
	}
	defer db.Close()

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(dbCtx, "INSERT INTO cotacoes (bid, create_at) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("erro ao preparar inserção no banco de dados: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(dbCtx, data, time.Now())
	if err != nil {
		return fmt.Errorf("erro ao gravar as informações no banco de dados: %w", err)
	}
	return nil
}
