package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func printSummary(totalAdapters int, keywords []string, jobsCount int, duration time.Duration) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 RELATÓRIO DE EXECUÇÃO DO SCRAPER")
	fmt.Println(strings.Repeat("=", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "MÉTRICA\t VALOR")
	fmt.Fprintln(w, "-------\t -----")
	fmt.Fprintf(w, "Adapters Carregados\t %d\n", totalAdapters)
	fmt.Fprintf(w, "Keywords Processadas\t %d\n", len(keywords))
	fmt.Fprintf(w, "Lista de Keywords\t %s\n", strings.Join(keywords, ", "))
	fmt.Fprintf(w, "Total Vagas (Pós-Dedup)\t %d\n", jobsCount)
	fmt.Fprintf(w, "Tempo Total de Resposta\t %v\n", duration.Round(time.Millisecond))
	w.Flush()

	fmt.Println(strings.Repeat("=", 60) + "\n")
}
