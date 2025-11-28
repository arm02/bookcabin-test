# Flight Search & Aggregation System (BookCabin Take-Home Test)

Ini adalah implementasi dari sistem agregasi pencarian penerbangan berbasis Go, dirancang dengan arsitektur modular yang memisahkan concerns dan mengutamakan performa API.

## Technical Design Choices

1.  **Clean Separation of Concerns (SoC):**

    - **Domain (`internal/core/domain`)**: Hanya berisi struct data, definisi model (UnifiedFlight, SearchCriteria, dll.), dan tipe data.
    - **Platform (`internal/platform/providers`)**: Berisi logika integrasi (API mocking) dan Normalisasi Data spesifik per maskapai (Lion Air, Batik Air, AirAsia, Garuda Indonesia).
    - **Service (`internal/core/services`)**: Berisi `Aggregator` sebagai inti bisnis. Logika Aggregator dibagi menjadi file terpisah:
      - `aggregator_fetch.go`: Concurrency & Error Handling
      - `aggregator_cache.go`: Caching Logic
      - `aggregator_filter.go`: Filtering Logic
      - `aggregator_sort.go`: Scoring & Sorting Logic
    - **Handler (`internal/handlers`)**: Layer ini berfungsi sebagai Antarmuka (Interface) antara World Wide Web (permintaan HTTP) dan logika inti bisnis (services)

2.  **API Performance & Concurrency:**

    - **Parallel Queries & Timeout**: Menggunakan `sync.WaitGroup` dan Goroutine untuk melakukan _queries_ ke semua provider secara simultan. _Aggregator_ menggunakan timeout total (500ms) pada proses _fetching_ untuk memastikan latensi total terkendali.
    - **Retry Logic dengan Exponential Backoff (Bonus)**: Untuk _provider_ yang flaky (disimulasikan pada AirAsia), diterapkan **Retry** Logic dengan **Exponential Backoff** (`2^n delay`) untuk mengurangi beban pada API provider saat terjadi kegagalan sementara.

3.  **Caching**

    - **Strategi**: Menggunakan Simple In-Memory Cache yang diimplementasikan dengan sync.Map untuk keamanan thread.
    - **Cache Key**: Kunci cache di-generate berdasarkan hash SHA256 dari kriteria pencarian inti (Origin, Destination, DepartureDate, CabinClass).
    - **Kedaluwarsa**: Data cache kedaluwarsa setelah 60 detik (CacheExpiration), memaksa re-fetch dari provider untuk memastikan kesegaran data harga.
    - **Filter pada Cache Hit**: Filter dan sorting diterapkan pada data yang di-cache saat terjadi Cache Hit untuk memastikan kriteria pencarian terbaru selalu dihormati.

4.  **Data Consistency & Error Handling:**

    - **Timezone & Data Validation**: Setiap provider menggunakan time.ParseInLocation untuk menangani konversi Timezone (WIB/WITA/WIT) yang benar. Data yang tidak valid (misalnya, Arrival Time sebelum Departure Time) ditandai dengan flag IsValid dan dihapus sebelum di-aggregate.
    - **"Best Value" Scoring Algorithm**: Algoritma ranking diimplementasikan untuk memberikan nilai kombinasi antara harga dan kenyamanan:

          $$\text{Score} = \left(\frac{\text{Price}}{1000}\right) + (\text{Duration Mins} \times 5) + (\text{Stops} \times 5000)$$

      Skor terendah adalah nilai terbaik.

## Implemented Bonus Points

| Persyaratan Bonus                                | Status |
| ------------------------------------------------ | ------ |
| "Best Value" Scoring Algorithm                   | ✅     |
| Handle Timezone Conversions                      | ✅     |
| Parallel Provider Queries with Timeout           | ✅     |
| Implement Retry Logic with Exponential Backoff   | ✅     |
| Clear Separation of Concerns (SoC)               | ✅     |
| Caching (In-memory, time-based)                  | ✅     |
| Supported Sorting Options (Harga, Durasi, Waktu) | ✅     |

## Setup dan Usage

### 1. Requirements

- Go (Golang) 1.21+

### 2. Clone Repository

```bash
git clone git@github.com:arm02/bookcabin-test.git
```

### 3. Run the Server

Pastikan Anda berada di direktori `bookcabin-test` dan run command bawah.

```bash
go mod tidy
```

```bash
go run cmd/api/main.go
```

### 4. Contoh Penggunaan API (Request)

Sistem ini menerima kriteria pencarian dalam format JSON berikut.

**Endpoint:** POST /v1/search (contoh asumsi)

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "returnDate": "2025-12-19",
  "passengers": 1,
  "cabinClass": "economy",
  "filters": {
    "maxPrice": 2000000,
    "minPrice": 500000,
    "maxStops": 1,
    "airlines": ["AirAsia", "Garuda Indonesia", "Lion Air", "Batik Air"],
    "minDurationMinutes": null,
    "maxDurationMinutes": null,
    "minDepTime": null,
    "maxDepTime": null,
    "minArrTime": null,
    "maxArrTime": null
  },
  "sortBy": "best_value" //default 
}
```
**CURL**
```bash
curl --location 'http://localhost:8080/v1/search' \
--header 'Content-Type: application/json' \
--data '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "returnDate": "2025-12-20",
    "passengers": 10,
    "cabinClass": "economy",
    "filters": {
        "maxPrice": 5000000,
        "minPrice": 500000,
        "maxStops": 1,
        "airlines": [
            "AirAsia",
            "Garuda Indonesia",
            "Lion Air",
            "Batik Air"
        ],
        "minDurationMinutes": null,
        "maxDurationMinutes": null,
        "minDepTime": null,
        "maxDepTime": null,
        "minArrTime": null,
        "maxArrTime": null
    },
    "sortBy": "best_value"
}'
```

| Field sortBy | Keterangan                                       |
| ------------ | ------------------------------------------------ |
| best_value   | (Default) Skor terendah adalah terbaik.          |
| price_asc    | Harga terendah ke tertinggi.                     |
| duration_asc | Durasi terpendek ke terlama.                     |
| dep_time_asc | Waktu keberangkatan paling pagi ke paling malam. |
| arr_time_asc | Waktu kedatangan paling pagi ke paling malam.    |

**Output Response**

```json
{
  "search_criteria": {
    /* ... kriteria yang digunakan ... */
  },
  "metadata": {
    "total_results": 8,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 115,
    "cache_hit": true
  },
  "flights": [
    {
      "id": "QZ532_AirAsia",
      "provider": "AirAsia",
      "airline": { "name": "AirAsia", "code": "QZ" },
      "flight_number": "QZ532",
      "stops": 0,
      "departure": {
        /* ... */
      },
      "arrival": {
        /* ... */
      },
      "duration": { "total_minutes": 160, "formatted": "2h 40m" },
      "price": { "amount": 595000, "currency": "IDR" },
      "available_seats": 72,
      "cabin_class": "economy",
      "score": 6092.345 // Digunakan untuk sorting "best_value"
    }
  ]
}
```
