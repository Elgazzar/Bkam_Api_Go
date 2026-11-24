# API Documentation: Dollar Price Egypt (Dollar_Price.apks)

## 1. Overview & Architecture

- **Application Name:** Dollar Price Egypt (سعر الدولار اليوم في مصر)
- **Package Name:** `com.dollarbankegypt`
- **Source File:** `Dollar_Price.apks`
- **Application Version:** `11.0.1` (Build `100`)
- **Technology Stack:** **Flutter** (Dart AOT compiled to native ARM64 assembly)
- **Backend Infrastructure:** **Supabase** (Managed PostgreSQL database exposed via PostgREST and Realtime WebSockets)

---

## 2. API Infrastructure & Credentials

The mobile app communicates directly with a hosted Supabase project.

| Parameter | Value |
| :--- | :--- |
| **Project URL** | `https://bqpaccgbujenwldzzivp.supabase.co` |
| **PostgREST REST API Base** | `https://bqpaccgbujenwldzzivp.supabase.co/rest/v1` |
| **Realtime Gateway** | `https://bqpaccgbujenwldzzivp.supabase.co/realtime/v1` |
| **Storage CDN (Logos)** | `https://bqpaccgbujenwldzzivp.supabase.co/storage/v1/object/public/bank-logos/` |
| **Anon Public Key (JWT)** | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8` |
| **Key Expiration** | Year **2036** (`exp: 2096494940`) |
| **Token Role** | `anon` |

---

## 3. Where the API Key is Stored in the APK

The API key and endpoint URLs are not stored in standard Android XML resource files or `.env` configuration files. Instead, they are compiled directly into the Flutter native binary snapshot:

```
Dollar_Price.apks
 └── split_config.arm64_v8a.apk
      └── lib/arm64-v8a/libapp.so
           └── ELF Section: .rodata
                └── Dart AOT Symbol: _kDartIsolateSnapshotData
                     └── File Offset: 0x2cbac (183,212 bytes)
```

### Exact Binary Location Details:
- **Archive Path:** `Dollar_Price.apks` -> `split_config.arm64_v8a.apk`
- **Binary File:** `lib/arm64-v8a/libapp.so`
- **ELF Section:** `.rodata` (Read-Only Data Segment)
- **Dart Symbol:** `_kDartIsolateSnapshotData` (Dart Isolate static string heap)
- **Absolute File Offset:** `0x2cbac` (Decimal: `183,212`)
- **Relative Offset in Snapshot:** `0x285ec` bytes into the isolate data pool

---

## 4. Authentication Methods

Supabase PostgREST endpoints support two authentication modes:

### A. Direct Browser URL (Query Parameter)
Pass the API key directly via the `apikey` query parameter. This allows testing directly in any standard browser address bar without setting custom headers:
```text
https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/<TABLE_NAME>?apikey=<ANON_KEY>
```

### B. Standard HTTP Headers (For Backend / Server Clients)
```http
apikey: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8
Accept: application/json
```

---

## 5. Endpoints & Schemas

### 5.1 Live Bank Dollar Rates (`banks`)
Retrieves real-time USD buy and sell rates across 25 Egyptian banks.

- **Direct Browser URL (Sorted by highest buy rate):**  
  [https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/banks?select=*&order=buy_rate.desc&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8](https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/banks?select=*&order=buy_rate.desc&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8)

- **Single Bank Filter (e.g., HSBC Egypt, ID `253`):**  
  [https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/banks?id=eq.253&select=*&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8](https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/banks?id=eq.253&select=*&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8)

#### Response Schema:
```json
{
  "id": 253,
  "name_ar": "اتش اس بي سي (HSBC)",
  "name_en": "اتش اس بي سي (HSBC)",
  "buy_rate": 54.15,
  "sell_rate": 54.25,
  "previous_buy_rate": 54.18,
  "previous_sell_rate": 54.28,
  "logo_url": "https://bqpaccgbujenwldzzivp.supabase.co/storage/v1/object/public/bank-logos/hsbc.png",
  "last_updated": "2026-11-08T09:34:01.634+00:00",
  "is_active": true,
  "sort_order": 99,
  "type": "dollar"
}
```

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | integer | Unique identifier of the bank |
| `name_ar` | string | Arabic bank name |
| `name_en` | string | English/transliterated bank name |
| `buy_rate` | number | Current bank buying rate in EGP |
| `sell_rate` | number | Current bank selling rate in EGP |
| `previous_buy_rate` | number | Buy rate prior to the latest update |
| `previous_sell_rate` | number | Sell rate prior to the latest update |
| `logo_url` | string | CDN URL to the bank's icon |
| `last_updated` | timestamp | ISO 8601 UTC timestamp of last rate refresh |
| `is_active` | boolean | Indicates whether the bank is currently active |
| `sort_order` | integer | Default display sort index |
| `type` | string | Feed type (`dollar`) |

---

### 5.2 Live Gold Prices (`gold_prices`)
Retrieves live Egyptian gold prices per karat, gold pound, ounce, and gold bars.

- **Direct Browser URL:**  
  [https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/gold_prices?select=*&order=id.asc&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8](https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/gold_prices?select=*&order=id.asc&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8)

#### Sample Response:
```json
[
  { "id": 1, "karat": "24", "name_ar": "ذهب عيار 24", "buy_price": 7091, "sell_price": 7020 },
  { "id": 2, "karat": "21", "name_ar": "ذهب عيار 21", "buy_price": 6205, "sell_price": 6184 },
  { "id": 3, "karat": "18", "name_ar": "ذهب عيار 18", "buy_price": 5318, "sell_price": 5265 },
  { "id": 4, "karat": "14", "name_ar": "ذهب عيار 14", "buy_price": 4137, "sell_price": 4096 },
  { "id": 5, "karat": "pound", "name_ar": "جنيه الذهب", "buy_price": 49640, "sell_price": 49144 },
  { "id": 6, "karat": "ounce", "name_ar": "أونصة الذهب", "buy_price": 220554.91, "sell_price": 218346.57 },
  { "id": 7, "karat": "bar_2.5", "name_ar": "سبيكة 2.5 جرام", "buy_price": 17727.5, "sell_price": 17550 },
  { "id": 8, "karat": "bar_5", "name_ar": "سبيكة 5 جرام", "buy_price": 35455, "sell_price": 35100 },
  { "id": 9, "karat": "bar_10", "name_ar": "سبيكة 10 جرام", "buy_price": 70910, "sell_price": 70200 }
]
```

---

### 5.3 Historical Bank Rate Changes (`rate_history`)
Maintains an audit trail of every price change for each bank.

- **Direct Browser URL (Latest 20 rate changes for HSBC Egypt ID `253`):**  
  [https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/rate_history?bank_id=eq.253&select=*&order=recorded_at.desc&limit=20&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8](https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/rate_history?bank_id=eq.253&select=*&order=recorded_at.desc&limit=20&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8)

#### Sample Response:
```json
[
  {
    "id": 88523,
    "bank_id": 253,
    "buy_rate": 54.15,
    "sell_rate": 54.25,
    "recorded_at": "2026-11-08T09:34:01.634+00:00"
  }
]
```

---

### 5.4 Gold Price History (`gold_price_history`)
Maintains historical records of gold prices per karat.

- **Direct Browser URL (Latest 20 updates for 21K Gold):**  
  [https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/gold_price_history?karat=eq.21&select=*&order=recorded_at.desc&limit=20&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8](https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/gold_price_history?karat=eq.21&select=*&order=recorded_at.desc&limit=20&apikey=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8)

---

## 6. Complete 25-Bank Mapping Directory

| Bank ID | Bank Name (Arabic) | Short Name / Code |
| :---: | :--- | :--- |
| `86` | أبوظبي الإسلامي (ADIB) | Abu Dhabi Islamic Bank |
| `149` | الكويت الوطني (NBK) | National Bank of Kuwait |
| `63` | بنك مصر (BM) | Banque Misr |
| `82` | بنك نكست (NXT) | Bank NXT |
| `101` | قناة السويس (SCB) | Suez Canal Bank |
| `38755` | الأهلي الكويتي (ABK) | Al Ahli Bank of Kuwait |
| `88` | البركة (Al Baraka) | Al Baraka Bank |
| `106` | العربي الأفريقي (AAIB) | Arab African International Bank |
| `129` | بنك بيت التمويل الكويتي (KFH) | Kuwait Finance House |
| `1309` | كريدي أجريكول (CA) | Credit Agricole |
| `38739` | بنك فيصل (Faisal) | Faisal Islamic Bank |
| `147` | التنمية الصناعية (IDB) | Industrial Development Bank |
| `107` | أبوظبي اﻷول (FABMISR) | First Abu Dhabi Bank Misr |
| `152` | المصرف المتحد (UB) | The United Bank |
| `72` | بنك الشركة المصرفية (SAIB) | SAIB Bank |
| `253` | اتش اس بي سي (HSBC) | HSBC Bank Egypt |
| `64` | بنك الإسكندرية (ALEXBANK) | Bank of Alexandria |
| `38753` | قطر الوطني (QNB Alahli) | QNB Alahli |
| `1308` | الأهلي المصري (NBE) | National Bank of Egypt |
| `60` | المصرف العربي (AIB) | Arab International Bank |
| `90` | التجاري الدولي (CIB) | Commercial International Bank |
| `105` | العقاري المصري العربي (EALB) | Egyptian Arab Land Bank |
| `65` | المصري الخليجي (EG Bank) | Egyptian Gulf Bank |
| `38754` | أبوظبي التجاري (ADCB) | Abu Dhabi Commercial Bank |
| `96` | التعمير والإسكان (HDB) | Housing & Development Bank |

---

## 7. Implementation Code Examples

### Python (Standard Library)
```python
import urllib.request
import json

BASE_URL = "https://bqpaccgbujenwldzzivp.supabase.co/rest/v1"
API_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8"

url = f"{BASE_URL}/banks?select=*&order=buy_rate.desc"
headers = {
    "apikey": API_KEY,
    "Authorization": f"Bearer {API_KEY}",
    "Accept": "application/json"
}

req = urllib.request.Request(url, headers=headers)
with urllib.request.urlopen(req) as response:
    banks = json.loads(response.read().decode("utf-8"))
    for b in banks:
        print(f"[{b['id']}] {b['name_ar']}: Buy {b['buy_rate']} EGP | Sell {b['sell_rate']} EGP")
```

### JavaScript / Node.js (Fetch)
```javascript
const BASE_URL = 'https://bqpaccgbujenwldzzivp.supabase.co/rest/v1';
const API_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8';

async function fetchRates() {
  const response = await fetch(`${BASE_URL}/banks?select=*&order=buy_rate.desc`, {
    headers: {
      'apikey': API_KEY,
      'Authorization': `Bearer ${API_KEY}`,
      'Accept': 'application/json'
    }
  });
  const data = await response.json();
  console.log(data);
}

fetchRates();
```

### cURL
```bash
curl -X GET "https://bqpaccgbujenwldzzivp.supabase.co/rest/v1/banks?select=*&order=buy_rate.desc" \
     -H "apikey: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8" \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJxcGFjY2didWplbndsZHp6aXZwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA5MTg5NDAsImV4cCI6MjA5NjQ5NDk0MH0.mawPac5_McHMwZFMcSkvfsCfo1I0-f81WktKjJIBRd8"
```
