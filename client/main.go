package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Currency struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Erro ao criar requisição:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer requisição ao servidor:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler resposta do servidor:", err)
		return
	}

	if len(body) == 0 {
		fmt.Println("a resposta da API está vazia.")
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Erro ao obter cotação:", string(body))
		return
	}

	var result Currency
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Erro ao processar a resposta JSON:", err)
		return
	}

	bid := result.USDBRL.Bid

	err = saveCotacaoToFile(bid)
	if err != nil {
		fmt.Println("Erro ao salvar a cotação no arquivo:", err)
	}
}

func saveCotacaoToFile(bid string) error {
	content := fmt.Sprintf("Dólar: %s\n", bid)

	err := os.WriteFile("cotacao.txt", []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo: %w", err)
	}

	fmt.Println("Cotação salva no arquivo cotacao.txt com sucesso!")
	return nil
}
