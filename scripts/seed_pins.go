package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/opsnerve/fireline/internal/auth"
)

func main() {
	hash, err := auth.HashPIN("123456")
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash error: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("docker", "exec", "-i", "fireline-postgres-1",
		"psql", "-U", "fireline", "-d", "fireline",
		"-c", fmt.Sprintf("UPDATE employees SET pin_hash = '%s' WHERE status = 'active'", hash))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "psql error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All active employees now have PIN: 123456")
}
