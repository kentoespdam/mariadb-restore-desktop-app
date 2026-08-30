// Debug: replay the partial stream manually.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/core/scanner"
)

func main() {
	dumpPath := "/tmp/dump-shop.sql"

	// Re-dump shop fresh.
	cmd := exec.Command("mariadb-dump",
		"-h", "127.0.0.1", "-P", "3307",
		"-u", "root", "-ptestpass",
		"--databases", "shop",
	)
	f, _ := os.Create(dumpPath)
	cmd.Stdout = f
	cmd.Run()
	f.Close()

	objs, _ := scanner.New().Scan(dumpPath)
	for _, o := range objs {
		fmt.Printf("%-15s %s.%s [%d,%d)\n", o.ObjectType, o.DatabaseName, o.ObjectName, o.StartByte, o.EndByte)
	}

	// Build the partial stream for products
	var products []scanner.Object
	for _, o := range objs {
		if o.DatabaseName == "shop" && o.ObjectName == "products" {
			products = append(products, o)
		}
	}
	sort.Slice(products, func(i, j int) bool { return products[i].StartByte < products[j].StartByte })

	src, _ := os.Open(dumpPath)
	defer src.Close()
	_, _ = src.Stat()

	header := "USE `shop`;\nSET FOREIGN_KEY_CHECKS=0;\nSET UNIQUE_CHECKS=0;\nSET NAMES utf8mb4;\n"
	footer := "SET FOREIGN_KEY_CHECKS=1;\nCOMMIT;\n"

	// We can't use streamer.Build from here (private). Just write the
	// file manually.
	out, _ := os.Create("/tmp/partial-stream.sql")
	defer out.Close()
	out.WriteString(header)
	for _, p := range products {
		_, err := io.CopyN(out, io.NewSectionReader(src, p.StartByte, p.EndByte-p.StartByte), p.EndByte-p.StartByte)
		if err != nil {
			fmt.Println("copy err:", err)
		}
	}
	out.WriteString(footer)
	out.Close()
	fmt.Println("\nwrote /tmp/partial-stream.sql")

	// Now apply to a target DB
	exec.Command("mariadb", "-h", "127.0.0.1", "-P", "3307", "-u", "root", "-ptestpass",
		"-e", "DROP DATABASE IF EXISTS target;").Run()
	exec.Command("mariadb", "-h", "127.0.0.1", "-P", "3307", "-u", "root", "-ptestpass",
		"-e", "CREATE DATABASE target;").Run()
	partial, _ := os.Open("/tmp/partial-stream.sql")
	defer partial.Close()
	apply := exec.Command("mariadb", "-h", "127.0.0.1", "-P", "3307", "-u", "root", "-ptestpass", "target")
	apply.Stdin = partial
	out2, _ := apply.CombinedOutput()
	fmt.Println("apply output:", string(out2))

	cnt, _ := exec.Command("mariadb", "-h", "127.0.0.1", "-P", "3307", "-u", "root", "-ptestpass",
		"target", "-N", "-B", "-e", "SELECT COUNT(*) FROM products;").Output()
	fmt.Println("target.products count:", string(cnt))
}