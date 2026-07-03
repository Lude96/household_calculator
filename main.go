package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/jung-kurt/gofpdf"
)

type Transaction struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"` // "income" or "expense"
	Notes    string  `json:"notes"`
}

type HouseholdData struct {
	Transactions []Transaction `json:"transactions"`
}

const dataFile = "household_data.json"

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addType := addCmd.String("type", "", "Type: income or expense")
	addCategory := addCmd.String("category", "", "Category name")
	addAmount := addCmd.Float64("amount", 0, "Amount")
	addNotes := addCmd.String("notes", "", "Notes")

	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportFormat := reportCmd.String("format", "console", "Format: console, pdf, or chart")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if *addType == "" || *addCategory == "" || *addAmount == 0 {
			if *addType == "" && *addCategory == "" && *addAmount == 0 {
				interactiveAdd()
				return
			}
			addCmd.PrintDefaults()
			return
		}
		addTransaction(*addType, *addCategory, *addAmount, *addNotes)

	case "report":
		reportCmd.Parse(os.Args[2:])
		generateReport(*reportFormat)

	case "list":
		listTransactions()

	case "clear":
		clearData()

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Household Calculator - Ausgaben- und Einnahmenverwaltung

Usage:
  household add [options]           - Transaktion hinzufügen
  household report [options]        - Report generieren
  household list                    - Alle Transaktionen anzeigen
  household clear                   - Alle Daten löschen

Add options:
  -type      income|expense
  -category  Kategorie
  -amount    Betrag
  -notes     Notizen (optional)

Report options:
  -format    console|pdf|chart (default: console)

Examples:
  household add -type expense -category Groceries -amount 50.25 -notes "Wochenmarkt"
  household report -format pdf
  household list`)
}

func interactiveAdd() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Type (income/expense): ")
	txType, _ := reader.ReadString('\n')
	txType = strings.TrimSpace(txType)

	fmt.Print("Category: ")
	category, _ := reader.ReadString('\n')
	category = strings.TrimSpace(category)

	fmt.Print("Amount: ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Println("Invalid amount:", err)
		return
	}

	fmt.Print("Notes (optional): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	addTransaction(txType, category, amount, notes)
}

func addTransaction(txType, category string, amount float64, notes string) {
	data := loadData()

	if txType != "income" && txType != "expense" {
		log.Println("Error: Type must be 'income' or 'expense'")
		return
	}

	transaction := Transaction{
		Category: category,
		Amount:   amount,
		Type:     txType,
		Notes:    notes,
	}

	data.Transactions = append(data.Transactions, transaction)
	saveData(data)
	fmt.Printf("✓ Transaction added: %s - %s: %.2f\n", txType, category, amount)
}

func listTransactions() {
	data := loadData()

	if len(data.Transactions) == 0 {
		fmt.Println("No transactions found.")
		return
	}

	fmt.Println("\n=== Transactions ===")
	totalIncome := 0.0
	totalExpense := 0.0

	for i, tx := range data.Transactions {
		icon := "➕"
		if tx.Type == "expense" {
			icon = "➖"
		}
		fmt.Printf("%d. %s %s - %s: %.2f", i+1, icon, tx.Type, tx.Category, tx.Amount)
		if tx.Notes != "" {
			fmt.Printf(" (%s)", tx.Notes)
		}
		fmt.Println()

		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpense += tx.Amount
		}
	}

	fmt.Printf("\nTotal Income: %.2f\n", totalIncome)
	fmt.Printf("Total Expenses: %.2f\n", totalExpense)
	fmt.Printf("Balance: %.2f\n", totalIncome-totalExpense)
}

func generateReport(format string) {
	data := loadData()

	if len(data.Transactions) == 0 {
		fmt.Println("No transactions to report.")
		return
	}

	switch format {
	case "console":
		listTransactions()
	case "pdf":
		generatePDF(data)
	case "chart":
		generateChart(data)
	default:
		fmt.Println("Unknown format:", format)
	}
}

func generateChart(data HouseholdData) {
	expenseMap := make(map[string]float64)
	totalIncome := 0.0
	totalExpense := 0.0

	for _, tx := range data.Transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			expenseMap[tx.Category] += tx.Amount
			totalExpense += tx.Amount
		}
	}

	pie := charts.NewPie()

	items := make([]opts.PieData, 0)
	for category, amount := range expenseMap {
		items = append(items, opts.PieData{Name: category, Value: amount})
	}

	// Add balance
	remaining := totalIncome - totalExpense
	if remaining > 0 {
		items = append(items, opts.PieData{Name: "Überschuss", Value: remaining})
	}

	pie.AddSeries("expenses", items)
	
	f, err := os.Create("expenses_chart.html")
	if err != nil {
		log.Println("Error creating chart:", err)
		return
	}
	defer f.Close()
	
	pie.Render(f)
	fmt.Printf("✓ Chart saved as expenses_chart.html\n  Budget: %.2f EUR | Ausgaben: %.2f EUR | Überschuss: %.2f EUR\n", 
		totalIncome, totalExpense, remaining)
}

func generatePDF(data HouseholdData) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Household Calculator Report", "", 1, "C", false, 0, "")

	totalIncome := 0.0
	totalExpense := 0.0
	expenseByCategory := make(map[string]float64)

	for _, tx := range data.Transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
		} else {
			totalExpense += tx.Amount
			expenseByCategory[tx.Category] += tx.Amount
		}
	}

	pdf.SetFont("Arial", "", 12)
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, "Summary", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Income: %.2f EUR", totalIncome), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Expenses: %.2f EUR", totalExpense), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Balance: %.2f EUR", totalIncome-totalExpense), "", 1, "L", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, "Expenses by Category", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	for category, amount := range expenseByCategory {
		pdf.CellFormat(0, 6, fmt.Sprintf("  %s: %.2f EUR", category, amount), "", 1, "L", false, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, "All Transactions", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, tx := range data.Transactions {
		icon := "Income"
		if tx.Type == "expense" {
			icon = "Expense"
		}
		pdf.CellFormat(0, 5, fmt.Sprintf("  [%s] %s: %.2f EUR", icon, tx.Category, tx.Amount), "", 1, "L", false, 0, "")
		if tx.Notes != "" {
			pdf.CellFormat(0, 4, fmt.Sprintf("    Note: %s", tx.Notes), "", 1, "L", false, 0, "")
		}
	}

	pdf.OutputFileAndClose("report.pdf")
	fmt.Println("✓ PDF report saved as report.pdf")
}

func loadData() HouseholdData {
	data := HouseholdData{}

	content, err := ioutil.ReadFile(dataFile)
	if err != nil {
		// File doesn't exist yet, return empty data
		return data
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Println("Error reading data:", err)
		return data
	}

	return data
}

func saveData(data HouseholdData) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("Error marshaling data:", err)
		return
	}

	err = ioutil.WriteFile(dataFile, jsonData, 0644)
	if err != nil {
		log.Println("Error saving data:", err)
	}
}

func clearData() {
	err := os.Remove(dataFile)
	if err != nil && !os.IsNotExist(err) {
		log.Println("Error clearing data:", err)
		return
	}
	fmt.Println("All data cleared.")
}
