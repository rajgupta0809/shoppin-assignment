package storage

import (
	"encoding/csv"
	"os"
)

func SaveToCSV(data []ProductData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Domain", "Product URL"})

	for _, entry := range data {
		for _, url := range entry.ProductURLs {
			writer.Write([]string{entry.Domain, url})
		}
	}
	return nil
}
