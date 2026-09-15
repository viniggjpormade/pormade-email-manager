package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	root := os.Getenv("APP_ROOT")
	if root == "" {
		_, currentFile, _, ok := runtime.Caller(0)
		if ok {
			root = filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
		} else {
			root, _ = os.Getwd()
		}
	}

	envPath := filepath.Join(root, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		candidate := filepath.Join(cwd, ".env")
		if _, err := os.Stat(candidate); err == nil {
			envPath = candidate
		}
	}

	if err := godotenv.Load(envPath); err != nil {
		fmt.Println("Aviso: Arquivo .env não encontrado. Utilizando variáveis injetadas pelo sistema/Docker.")
	}
}
