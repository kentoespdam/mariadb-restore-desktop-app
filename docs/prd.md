Sempurna! Semua 5 poin krusial arsitektur telah disepakati bersama. Kita berhasil menyusun arsitektur pengganti **MySQL Administrator** untuk **MariaDB 10.11+** yang tangguh, hemat memori, dan *fully portable*.

Berikut adalah **Technical Architecture Blueprint** lengkap berdasarkan hasil *grilling* kita:

---

## 🛠️ Technical Blueprint: Portable MariaDB Restore & Backup Tool (Go + Wails)

```
 [ Dump File (.sql) ]
         │
         ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │ 1. Fast Byte-Offset Scanner (Go bufio.Reader)                          │
 └───────┬────────────────────────────────────────────────────────────────┘
         │ Populates Index
         ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │ 2. Portable Catalog (Embedded SQLite + AES-256-GCM app.key)           │
 └───────┬────────────────────────────────────────────────────────────────┘
         │ User selects tables via Wails UI (Virtual Scroll Grid)
         ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │ 3. Virtual Streamer (io.MultiReader)                                   │
 │    - Header Stream (Session Flags)                                     │
 │    - Selected Bytes Stream (io.SectionReader)                          │
 │    - Footer Stream                                                     │
 └───────┬────────────────────────────────────────────────────────────────┘
         │ On-The-Fly Filtering (Definer Stripper / DEFINER=CURRENT_USER)
         ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │ 4. ProgressReader Wrapper (Atomic bytes + Throttled Wails Events)      │
 └───────┬────────────────────────────────────────────────────────────────┘
         │ Stdin Pipe
         ▼
 [ Subprocess: mariadb CLI (exec.CommandContext with Cancel Support) ]

```

---

### 1. Storage & Portable Security Layer

* **Executable Scope:** Seluruh file aplikasi disimpan selevel dengan binary (`os.Executable()`).
* **Auto-Generated Key (`app.key`):**
* Kunci acak 256-bit (AES-GCM) dibuat otomatis saat *first launch*.
* Digunakan untuk mengenkripsi kredensial server database yang tersimpan di SQLite lokal.


* **Smart Recovery Modal:**
* Jika `app.key` hilang namun SQLite lama ditemukan:
* **Action A (Cancel):** Menutup aplikasi agar pengguna dapat memulihkan `app.key` yang tertinggal.
* **Action B (Reset & Re-init):** Menghapus SQLite lama, membuat `app.key` baru, dan menginisialisasi ulang konfigurasi.





---

### 2. Dump Analysis & Byte-Offset Indexer

* **Single-Pass Scanning:**
* Membaca file `.sql` menggunakan `bufio.Reader` tanpa memuat seluruh file ke RAM.
* Mencatat koordinat `start_byte` dan `end_byte` untuk DDL (`CREATE TABLE`) dan DML (`INSERT INTO`) tiap objek ke dalam SQLite catalog.


* **Instant Filtering & UX:**
* UI Wails membaca daftar objek dari SQLite catalog lokal (pencarian & pencentangan tabel bersifat *instant*).
* Menggunakan **Virtual Scrolling** di Frontend Wails untuk performa render 60 FPS pada skema raksasa (>10.000 tabel).



---

### 3. Execution Engine & On-The-Fly Stream Filtering

* **Virtual MultiStream Pipeline:**
Menggabungkan 3 alur stream menggunakan `io.MultiReader`:
1. **Header Stream:** `SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET NAMES utf8mb4;`
2. **Selected Tables Stream:** Membaca rentang offset byte yang dipilih menggunakan `io.NewSectionReader`.
3. **Footer Stream:** `SET FOREIGN_KEY_CHECKS=1; COMMIT;`


* **Definer Stripper Transformer:**
* Custom `io.Reader` wrapper menggantikan klausa `DEFINER=` menjadi `DEFINER=CURRENT_USER` atau menghapusnya secara *real-time* saat stream mengalir ke CLI.


* **Direct CLI Pipe:**
* Stream dipipakan langsung ke `stdin` dari `mariadb` CLI subprocess (`os/exec`). Auto-discovery melacak lokasi `mariadb.exe` di Windows via Registry/PATH atau `/usr/bin/mariadb` di Linux.



---

### 4. Progress Tracking & Graceful Cancellation

* **Byte Progress Metrics:**
* Custom `ProgressReader` membungkus stream untuk menghitung persentase byte terkirim secara *atomic*.
* `time.Ticker` mengirimkan event progress ke UI Wails setiap 100ms–250ms untuk mencegah *flooding* IPC.


* **Instant Cancellation:**
* Pembatalan dibungkus menggunakan `context.WithCancel(ctx)`.
* Saat tombol Cancel ditekan di UI, Go memanggil `cancel()`, membunuh subprocess `mariadb` secara anggun, memutus `stdin`, dan menutup *file handle* tanpa meninggalkan *zombie process*.



---